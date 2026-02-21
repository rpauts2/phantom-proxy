"""
PhantomProxy AI Service v13.0 - LangGraph + Llama-3.1 + RAG
Enterprise AI for phishing campaign generation and analysis
"""
import os
import asyncio
import json
import hashlib
from typing import Dict, List, Any, Optional
from contextlib import asynccontextmanager
from datetime import datetime

from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
import httpx

# LangGraph imports (optional - use if available)
try:
    from langgraph.graph import StateGraph, END
    from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
    from langchain_core.embeddings import Embeddings
    LANGGRAPH_AVAILABLE = True
except ImportError:
    LANGGRAPH_AVAILABLE = False

# Llama/Ollama client
try:
    import ollama
    OLLAMA_AVAILABLE = True
except ImportError:
    OLLAMA_AVAILABLE = False

# Vector Store (optional)
try:
    import chromadb
    from chromadb.config import Settings as ChromaSettings
    VECTOR_STORE_AVAILABLE = True
except ImportError:
    VECTOR_STORE_AVAILABLE = False


# ============================================================================
# Configuration
# ============================================================================

class Settings(BaseModel):
    service_name: str = "phantomproxy-ai"
    llm_provider: str = "ollama"  # ollama, openai, anthropic
    llm_model: str = "llama3.1:70b"
    llm_endpoint: str = "http://localhost:11434"
    openai_api_key: Optional[str] = None
    openai_model: str = "gpt-4"
    max_tokens: int = 4096
    temperature: float = 0.7
    rag_enabled: bool = False
    vector_store_path: str = "./vector_store"

    class Config:
        env_file = ".env"


settings = Settings()


# ============================================================================
# Request/Response Models
# ============================================================================

class GenerateEmailRequest(BaseModel):
    target_data: Dict[str, Any] = Field(..., description="Target information")
    template: str = Field(..., description="Email template type")
    language: str = Field(default="en", description="Response language")
    tone: str = Field(default="professional", description="Email tone")


class GenerateEmailResponse(BaseModel):
    success: bool
    email_body: Optional[str] = None
    subject: Optional[str] = None
    suggestions: List[str] = []
    confidence: float = 0.0


class PersonalizeRequest(BaseModel):
    content: str
    target_profile: Dict[str, Any]


class AnalyzeCredentialRequest(BaseModel):
    username: str
    password: Optional[str] = None
    metadata: Dict[str, Any] = {}


class AnalyzeCredentialResponse(BaseModel):
    success: bool
    analysis: Dict[str, Any]
    recommendations: List[str]
    risk_score: float = 0.0


class GenerateReportRequest(BaseModel):
    campaign_id: str
    campaign_data: Dict[str, Any]
    report_type: str = "executive"  # executive, technical, detailed


class ChatMessage(BaseModel):
    role: str  # system, user, assistant
    content: str


class ChatRequest(BaseModel):
    messages: List[ChatMessage]
    system_prompt: Optional[str] = None


# ============================================================================
# AI Service State
# ============================================================================

class AIState(BaseModel):
    """State for LangGraph workflow"""
    target_data: Dict[str, Any] = {}
    template: str = ""
    generated_content: str = ""
    refined_content: str = ""
    suggestions: List[str] = []
    confidence: float = 0.0
    errors: List[str] = []


# ============================================================================
# RAG Vector Store Manager
# ============================================================================

class RAGVectorStore:
    """Manages vector store for RAG-based AI generation"""

    def __init__(self, path: str = "./vector_store"):
        self.path = path
        self.client = None
        self.collection = None
        self._init_store()

    def _init_store(self):
        """Initialize vector store"""
        if not VECTOR_STORE_AVAILABLE:
            return

        try:
            self.client = chromadb.Client(ChromaSettings(
                persist_directory=self.path,
                anonymized_telemetry=False
            ))
            self.collection = self.client.get_or_create_collection(
                name="phantom_knowledge",
                metadata={"description": "Phishing campaign knowledge base"}
            )
        except Exception as e:
            print(f"Vector store init error: {e}")

    def add_document(self, doc_id: str, content: str, metadata: Dict[str, Any] = None):
        """Add document to vector store"""
        if not self.collection:
            return

        try:
            self.collection.upsert(
                ids=[doc_id],
                documents=[content],
                metadatas=[metadata or {}]
            )
        except Exception as e:
            print(f"Add document error: {e}")

    def search(self, query: str, n_results: int = 5) -> List[Dict[str, Any]]:
        """Search for relevant documents"""
        if not self.collection:
            return []

        try:
            results = self.collection.query(
                query_texts=[query],
                n_results=n_results
            )

            if results and results['documents']:
                return [
                    {
                        "content": doc,
                        "metadata": results['metadatas'][0][i] if results['metadatas'] else {},
                        "distance": results['distances'][0][i] if results.get('distances') else 0
                    }
                    for i, doc in enumerate(results['documents'][0])
                ]
            return []
        except Exception as e:
            print(f"Search error: {e}")
            return []

    def get_stats(self) -> Dict[str, Any]:
        """Get vector store statistics"""
        if not self.collection:
            return {"enabled": False}

        try:
            count = self.collection.count()
            return {
                "enabled": True,
                "documents": count,
                "collection": self.collection.name
            }
        except Exception as e:
            return {"enabled": False, "error": str(e)}


# ============================================================================
# LLM Client
# ============================================================================

class LLMClient:
    def __init__(self, settings: Settings, vector_store: Optional[RAGVectorStore] = None):
        self.settings = settings
        self.vector_store = vector_store
        self.client = None
        self._init_client()

    def _init_client(self):
        """Initialize LLM client based on provider"""
        if self.settings.llm_provider == "ollama" and OLLAMA_AVAILABLE:
            self.client = ollama.Client(host=self.settings.llm_endpoint)
        elif self.settings.llm_provider == "openai" and self.settings.openai_api_key:
            from openai import OpenAI
            self.client = OpenAI(api_key=self.settings.openai_api_key)

    async def generate(self, prompt: str, system_prompt: Optional[str] = None, use_rag: bool = True) -> str:
        """Generate text using configured LLM with optional RAG"""
        try:
            # RAG enhancement
            context = ""
            if use_rag and self.vector_store and self.settings.rag_enabled:
                rag_results = self.vector_store.search(prompt, n_results=3)
                if rag_results:
                    context = "\n\nRelevant context from knowledge base:\n"
                    context += "\n".join([r['content'][:500] for r in rag_results])

            full_prompt = prompt + context

            if self.settings.llm_provider == "ollama" and self.client:
                response = await asyncio.to_thread(
                    self.client.generate,
                    model=self.settings.llm_model,
                    prompt=full_prompt,
                    system=system_prompt or "You are a helpful AI assistant.",
                    stream=False
                )
                return response['response']

            elif self.settings.llm_provider == "openai" and self.client:
                response = await asyncio.to_thread(
                    self.client.chat.completions.create,
                    model=self.settings.openai_model,
                    messages=[
                        {"role": "system", "content": system_prompt or "You are a helpful AI assistant."},
                        {"role": "user", "content": full_prompt}
                    ],
                    max_tokens=self.settings.max_tokens,
                    temperature=self.settings.temperature
                )
                return response.choices[0].message.content

            else:
                # Fallback mock response
                return self._mock_response(prompt)

        except Exception as e:
            raise HTTPException(status_code=500, detail=f"LLM generation failed: {str(e)}")

    def _mock_response(self, prompt: str) -> str:
        """Mock response when no LLM is available"""
        return f"[Mock AI Response] Received prompt: {prompt[:100]}..."


# ============================================================================
# LangGraph Workflow (Optional)
# ============================================================================

def create_ai_workflow():
    """Create LangGraph workflow for email generation"""
    if not LANGGRAPH_AVAILABLE:
        return None

    def analyze_target(state: AIState) -> AIState:
        """Analyze target data"""
        # Extract key information
        state['suggestions'].append(f"Analyzing target: {state['target_data'].get('name', 'Unknown')}")
        return state

    def generate_draft(state: AIState) -> AIState:
        """Generate email draft"""
        state['generated_content'] = f"Draft email for {state['target_data'].get('name', 'target')}..."
        return state

    def refine_content(state: AIState) -> AIState:
        """Refine generated content"""
        state['refined_content'] = f"Refined: {state['generated_content']}"
        state['confidence'] = 0.85
        return state

    # Build workflow
    workflow = StateGraph(AIState)
    workflow.add_node("analyze", analyze_target)
    workflow.add_node("generate", generate_draft)
    workflow.add_node("refine", refine_content)

    workflow.set_entry_point("analyze")
    workflow.add_edge("analyze", "generate")
    workflow.add_edge("generate", "refine")
    workflow.add_edge("refine", END)

    return workflow.compile()


# ============================================================================
# FastAPI Application
# ============================================================================

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    vector_store = RAGVectorStore(settings.vector_store_path) if settings.rag_enabled else None
    app.state.llm_client = LLMClient(settings, vector_store)
    app.state.vector_store = vector_store
    app.state.workflow = create_ai_workflow()
    print(f"AI Service started with {settings.llm_provider}/{settings.llm_model}")
    print(f"RAG enabled: {settings.rag_enabled and VECTOR_STORE_AVAILABLE}")
    yield
    # Shutdown
    print("AI Service stopped")


app = FastAPI(
    title="PhantomProxy AI Service",
    description="Enterprise AI for phishing campaigns",
    version="13.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ============================================================================
# API Endpoints
# ============================================================================

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    vector_store_stats = app.state.vector_store.get_stats() if app.state.vector_store else {"enabled": False}

    return {
        "status": "healthy",
        "service": "phantomproxy-ai",
        "version": "13.0.0",
        "llm_provider": settings.llm_provider,
        "llm_model": settings.llm_model,
        "langgraph": LANGGRAPH_AVAILABLE,
        "ollama": OLLAMA_AVAILABLE,
        "vector_store": VECTOR_STORE_AVAILABLE,
        "rag_enabled": settings.rag_enabled,
        "rag_stats": vector_store_stats
    }


@app.get("/models")
async def list_models():
    """List available AI models"""
    return {
        "current": settings.llm_model,
        "available": [
            "llama3.1:70b",
            "llama3.1:8b",
            "mistral:7b",
            "mixtral:8x7b",
        ]
    }


@app.post("/v1/generate/email", response_model=GenerateEmailResponse)
async def generate_email(request: GenerateEmailRequest):
    """Generate phishing email using AI"""
    llm = app.state.llm_client

    # Build prompt
    system_prompt = f"""You are an expert at creating convincing phishing emails for security testing.
Create professional, personalized emails that pass security filters.
Language: {request.language}
Tone: {request.tone}"""

    prompt = f"""Generate a phishing email using the {request.template} template.

Target Information:
- Name: {request.target_data.get('name', 'Unknown')}
- Company: {request.target_data.get('company', 'Unknown')}
- Position: {request.target_data.get('position', 'Unknown')}
- Email: {request.target_data.get('email', 'Unknown')}
- Interests: {', '.join(request.target_data.get('interests', []))}

Create a convincing email that:
1. Appears legitimate and urgent
2. Uses personalization to build trust
3. Includes a clear call-to-action
4. Avoids common phishing indicators

Email:"""

    try:
        email_body = await llm.generate(prompt, system_prompt)

        # Generate subject line
        subject_prompt = f"Generate a compelling subject line for this email:\n\n{email_body[:500]}"
        subject = await llm.generate(subject_prompt, "Generate short, urgent subject lines.")

        return GenerateEmailResponse(
            success=True,
            email_body=email_body,
            subject=subject,
            suggestions=[
                "Personalize sender name",
                "Add company logo",
                "Test spam score"
            ],
            confidence=0.85
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/personalize", response_model=Dict[str, Any])
async def personalize_content(request: PersonalizeRequest):
    """Personalize content for specific target"""
    llm = app.state.llm_client

    prompt = f"""Personalize this content for the target:

Target Profile:
- Name: {request.target_profile.get('name', 'Unknown')}
- Company: {request.target_profile.get('company', 'Unknown')}
- Role: {request.target_profile.get('role', 'Unknown')}
- Interests: {', '.join(request.target_profile.get('interests', []))}

Original Content:
{request.content[:2000]}

Personalized Content:"""

    personalized = await llm.generate(prompt)

    return {
        "success": True,
        "original_length": len(request.content),
        "personalized_length": len(personalized),
        "content": personalized
    }


@app.post("/v1/analyze/credential", response_model=AnalyzeCredentialResponse)
async def analyze_credential(request: AnalyzeCredentialRequest):
    """Analyze captured credentials"""
    # Password strength analysis
    password = request.password or ""
    risk_score = 0.0
    recommendations = []

    if password:
        # Basic strength checks
        if len(password) < 8:
            risk_score += 0.3
            recommendations.append("Password too short")
        if not any(c.isupper() for c in password):
            risk_score += 0.2
            recommendations.append("Add uppercase letters")
        if not any(c.isdigit() for c in password):
            risk_score += 0.2
            recommendations.append("Add numbers")
        if not any(c in "!@#$%^&*()_+-=[]{}|;:,.<>?" for c in password):
            risk_score += 0.2
            recommendations.append("Add special characters")

    # Check for common passwords
    common_passwords = ["password", "123456", "qwerty", "admin", "letmein"]
    if password.lower() in common_passwords:
        risk_score += 0.5
        recommendations.append("Common password detected")

    # Breach check (mock - integrate with HIBP API)
    recommendations.append("Check Have I Been Pwned API")

    return AnalyzeCredentialResponse(
        success=True,
        analysis={
            "username": request.username,
            "password_length": len(password),
            "has_uppercase": any(c.isupper() for c in password),
            "has_lowercase": any(c.islower() for c in password),
            "has_numbers": any(c.isdigit() for c in password),
            "has_special": any(c in "!@#$%^&*()_+-=[]{}|;:,.<>?" for c in password),
        },
        recommendations=recommendations,
        risk_score=min(risk_score, 1.0)
    )


@app.post("/v1/report/generate")
async def generate_report(request: GenerateReportRequest):
    """Generate campaign report"""
    llm = app.state.llm_client

    prompt = f"""Generate a {request.report_type} security assessment report.

Campaign ID: {request.campaign_id}
Campaign Data:
{str(request.campaign_data)[:3000]}

Include:
1. Executive Summary
2. Key Findings
3. Statistics
4. Recommendations
5. Risk Assessment

Report:"""

    report = await llm.generate(prompt, "Write professional security assessment reports.")

    return {
        "success": True,
        "report_type": request.report_type,
        "campaign_id": request.campaign_id,
        "content": report,
        "generated_at": asyncio.get_event_loop().time()
    }


@app.post("/v1/chat")
async def chat(request: ChatRequest):
    """Chat with AI assistant"""
    llm = app.state.llm_client

    # Format messages
    messages_text = "\n".join([f"{m.role}: {m.content}" for m in request.messages])

    response = await llm.generate(
        messages_text,
        request.system_prompt or "You are a helpful security testing assistant."
    )

    return {
        "success": True,
        "response": response,
        "model": settings.llm_model
    }


@app.post("/v1/analyze/site")
async def analyze_site(url: str):
    """Analyze target website for phishing opportunities"""
    llm = app.state.llm_client

    prompt = f"""Analyze this website for phishing campaign opportunities:
URL: {url}

Provide:
1. Login page detection
2. Brand elements to replicate
3. Security measures detected
4. Recommended attack vectors
5. Phishlet configuration suggestions

Analysis:"""

    analysis = await llm.generate(prompt)

    return {
        "success": True,
        "url": url,
        "analysis": analysis,
        "recommendations": [
            "Create landing page clone",
            "Set up similar domain",
            "Configure SSL certificate"
        ]
    }


@app.get("/v1/stats")
async def get_stats():
    """Get AI service statistics"""
    vector_store_stats = app.state.vector_store.get_stats() if app.state.vector_store else {"enabled": False}

    return {
        "requests_processed": 0,  # Implement counter
        "avg_response_time": "0ms",
        "models_available": 4,
        "langgraph_enabled": LANGGRAPH_AVAILABLE,
        "ollama_enabled": OLLAMA_AVAILABLE,
        "vector_store_enabled": VECTOR_STORE_AVAILABLE,
        "rag_stats": vector_store_stats
    }


# ============================================================================
# RAG Knowledge Base Endpoints
# ============================================================================

class AddKnowledgeRequest(BaseModel):
    doc_id: str
    content: str
    metadata: Dict[str, Any] = {}
    doc_type: str = "campaign"  # campaign, phishlet, template, report


class SearchKnowledgeRequest(BaseModel):
    query: str
    n_results: int = 5
    doc_type: Optional[str] = None


@app.post("/v1/rag/add")
async def add_to_knowledge(request: AddKnowledgeRequest):
    """Add document to RAG knowledge base"""
    if not app.state.vector_store:
        raise HTTPException(status_code=400, detail="RAG not enabled")

    try:
        app.state.vector_store.add_document(
            doc_id=request.doc_id,
            content=request.content,
            metadata={**request.metadata, "type": request.doc_type, "added_at": datetime.now().isoformat()}
        )
        return {"success": True, "doc_id": request.doc_id}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/rag/search")
async def search_knowledge(request: SearchKnowledgeRequest):
    """Search knowledge base"""
    if not app.state.vector_store:
        raise HTTPException(status_code=400, detail="RAG not enabled")

    try:
        results = app.state.vector_store.search(request.query, request.n_results)

        # Filter by type if specified
        if request.doc_type:
            results = [r for r in results if r.get('metadata', {}).get('type') == request.doc_type]

        return {
            "success": True,
            "query": request.query,
            "results": results,
            "count": len(results)
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/v1/rag/stats")
async def rag_stats():
    """Get RAG knowledge base statistics"""
    if not app.state.vector_store:
        return {"enabled": False, "message": "RAG not enabled"}

    return {
        "success": True,
        "stats": app.state.vector_store.get_stats()
    }


# ============================================================================
# Phishlet Generator (AI-powered)
# ============================================================================

class GeneratePhishletRequest(BaseModel):
    target_url: str
    target_name: str
    login_fields: List[str] = ["username", "password"]
    additional_fields: List[str] = []


@app.post("/v1/generate/phishlet")
async def generate_phishlet(request: GeneratePhishletRequest):
    """Generate phishlet configuration using AI"""
    llm = app.state.llm_client

    system_prompt = """You are an expert at creating Evilginx phishlet configurations.
Generate valid YAML phishlet configurations for phishing simulations."""

    prompt = f"""Generate an Evilginx phishlet configuration for {request.target_name}.

Target URL: {request.target_url}
Login Fields: {', '.join(request.login_fields)}
Additional Fields: {', '.join(request.additional_fields)}

Create a complete phishlet YAML configuration including:
1. proxy_hosts configuration
2. sub_filters for domain replacement
3. auth_tokens for session capture
4. credentials configuration for field capture
5. auth_urls for login detection
6. js_inject for any required JavaScript

Output ONLY the YAML configuration:"""

    try:
        phishlet_yaml = await llm.generate(prompt, system_prompt, use_rag=True)

        return {
            "success": True,
            "target": request.target_name,
            "phishlet": phishlet_yaml,
            "note": "Review and test phishlet before deployment"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ============================================================================
# Main
# ============================================================================

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8081,
        reload=True,
        log_level="info"
    )
