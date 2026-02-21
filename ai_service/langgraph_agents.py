"""
PhantomProxy LangGraph Agents - Autonomous AI Workflows
Multi-agent system for phishing campaign automation
"""
import os
from typing import Dict, List, Any, Optional, Annotated
from dataclasses import dataclass, field
from datetime import datetime

# LangGraph imports
try:
    from langgraph.graph import StateGraph, END, StateSchema
    from langgraph.prebuilt import ToolNode
    from langchain_core.messages import HumanMessage, AIMessage, SystemMessage, BaseMessage
    from langchain_core.tools import tool
    from langchain_core.runnables import RunnableConfig
    LANGGRAPH_AVAILABLE = True
except ImportError:
    LANGGRAPH_AVAILABLE = False
    print("LangGraph not available - install with: pip install langgraph")

# LLM
try:
    import ollama
    OLLAMA_AVAILABLE = True
except ImportError:
    OLLAMA_AVAILABLE = False


# ============================================================================
# Agent State
# ============================================================================

@dataclass
class AgentState:
    """State for multi-agent workflow"""
    messages: List[BaseMessage] = field(default_factory=list)
    target_data: Dict[str, Any] = field(default_factory=dict)
    campaign_id: str = ""
    generated_content: Dict[str, str] = field(default_factory=dict)
    research_results: Dict[str, Any] = field(default_factory=dict)
    risk_assessment: Dict[str, float] = field(default_factory=dict)
    final_output: Dict[str, Any] = field(default_factory=dict)
    errors: List[str] = field(default_factory=list)
    current_step: str = ""
    completed_steps: List[str] = field(default_factory=list)


# ============================================================================
# Tools
# ============================================================================

@tool
def research_target(company: str, position: str) -> str:
    """Research target company and position for personalization"""
    # Mock - in production: integrate with LinkedIn, company website, etc.
    return f"""
    Company: {company}
    - Industry: Technology
    - Size: 1000-5000 employees
    - Recent news: Launched new cloud platform

    Position: {position}
    - Typical responsibilities: Team management, budget approval
    - Common tools: Email, CRM, ERP systems
    - Pain points: Security concerns, compliance requirements
    """


@tool
def analyze_email_pattern(email: str) -> dict:
    """Analyze email pattern for credibility"""
    parts = email.split('@')
    domain = parts[1] if len(parts) > 1 else ""

    corporate_domains = ['corp', 'company', 'enterprise', 'business', 'inc']
    is_corporate = any(d in domain.lower() for d in corporate_domains)

    return {
        "email": email,
        "domain": domain,
        "is_corporate": is_corporate,
        "credibility_score": 0.8 if is_corporate else 0.5
    }


@tool
def generate_subject_line(content: str, urgency: str = "medium") -> str:
    """Generate compelling subject line based on content"""
    urgency_words = {
        "low": "Update",
        "medium": "Important",
        "high": "URGENT",
        "critical": "IMMEDIATE ACTION REQUIRED"
    }

    prefix = urgency_words.get(urgency, "Important")
    return f"{prefix}: Account Verification Required"


@tool
def check_spam_score(content: str) -> dict:
    """Check content for spam indicators"""
    spam_words = ['free', 'winner', 'click here', 'act now', 'limited time']
    found = [word for word in spam_words if word.lower() in content.lower()]

    return {
        "spam_score": len(found) * 0.1,
        "flagged_words": found,
        "recommendation": "Remove spam trigger words" if found else "Content looks good"
    }


@tool
def personalize_content(template: str, target_data: dict) -> str:
    """Personalize template with target information"""
    content = template
    for key, value in target_data.items():
        content = content.replace(f"{{{{{key}}}}}", str(value))
    return content


@tool
def assess_risk_level(target_profile: dict) -> dict:
    """Assess phishing risk level for target"""
    score = 0.5  # Base score

    # Corporate email increases success probability
    if 'email' in target_profile:
        if any(d in target_profile['email'].lower() for d in ['corp', 'company']):
            score += 0.2

    # Senior position increases success probability
    if 'position' in target_profile:
        pos = target_profile['position'].lower()
        if any(title in pos for title in ['manager', 'director', 'head', 'chief']):
            score += 0.15

    return {
        "success_probability": min(score, 0.95),
        "risk_level": "high" if score > 0.7 else "medium" if score > 0.4 else "low",
        "recommended_approach": "spear_phishing" if score > 0.7 else "bulk_campaign"
    }


# ============================================================================
# Agent Nodes
# ============================================================================

class ResearchAgent:
    """Agent responsible for target research"""

    def __init__(self, llm_client=None):
        self.llm = llm_client
        self.tools = [research_target, analyze_email_pattern]

    def __call__(self, state: AgentState) -> AgentState:
        """Execute research step"""
        state.current_step = "research"

        company = state.target_data.get('company', 'Unknown')
        position = state.target_data.get('position', 'Unknown')
        email = state.target_data.get('email', '')

        # Use tools
        company_research = research_target.invoke({"company": company, "position": position})
        email_analysis = analyze_email_pattern.invoke({"email": email})

        state.research_results = {
            "company_info": company_research,
            "email_analysis": email_analysis,
            "timestamp": datetime.now().isoformat()
        }

        state.messages.append(AIMessage(content=f"Research completed for {company}"))
        state.completed_steps.append("research")

        return state


class ContentAgent:
    """Agent responsible for content generation"""

    def __init__(self, llm_client=None):
        self.llm = llm_client
        self.tools = [generate_subject_line, personalize_content, check_spam_score]

    def __call__(self, state: AgentState) -> AgentState:
        """Execute content generation step"""
        state.current_step = "content_generation"

        template = state.generated_content.get('template', '''
        Dear {{name}},

        We have detected unusual activity on your {{company}} account.

        Please verify your information immediately.

        Best regards,
        Security Team
        ''')

        # Personalize
        personalized = personalize_content.invoke({
            "template": template,
            "target_data": state.target_data
        })

        # Generate subject
        subject = generate_subject_line.invoke({
            "content": personalized,
            "urgency": "high"
        })

        # Check spam
        spam_check = check_spam_score.invoke({"content": personalized})

        state.generated_content = {
            "email_body": personalized,
            "subject_line": subject,
            "spam_score": spam_check['spam_score'],
            "template_used": "security_alert"
        }

        state.messages.append(AIMessage(content="Content generated successfully"))
        state.completed_steps.append("content_generation")

        return state


class RiskAgent:
    """Agent responsible for risk assessment"""

    def __init__(self, llm_client=None):
        self.llm = llm_client
        self.tools = [assess_risk_level]

    def __call__(self, state: AgentState) -> AgentState:
        """Execute risk assessment step"""
        state.current_step = "risk_assessment"

        risk = assess_risk_level.invoke({"target_profile": state.target_data})

        state.risk_assessment = {
            "success_probability": risk['success_probability'],
            "risk_level": risk['risk_level'],
            "recommended_approach": risk['recommended_approach'],
            "factors": {
                "email_quality": state.research_results.get('email_analysis', {}).get('credibility_score', 0.5),
                "position_seniority": 0.7 if 'manager' in str(state.target_data.get('position', '')).lower() else 0.4,
            }
        }

        state.messages.append(AIMessage(content=f"Risk assessed: {risk['risk_level']}"))
        state.completed_steps.append("risk_assessment")

        return state


class ReviewAgent:
    """Agent responsible for final review and output"""

    def __init__(self, llm_client=None):
        self.llm = llm_client

    def __call__(self, state: AgentState) -> AgentState:
        """Execute final review step"""
        state.current_step = "review"

        # Compile final output
        state.final_output = {
            "campaign_id": state.campaign_id,
            "target": state.target_data,
            "email": {
                "subject": state.generated_content.get('subject_line', ''),
                "body": state.generated_content.get('email_body', ''),
                "spam_score": state.generated_content.get('spam_score', 0),
            },
            "research": state.research_results,
            "risk": state.risk_assessment,
            "recommendations": [
                "Review content for brand compliance",
                "Test spam score with actual filters",
                "Verify all links are tracked"
            ],
            "generated_at": datetime.now().isoformat(),
            "steps_completed": state.completed_steps
        }

        state.messages.append(AIMessage(content="Review completed - campaign ready"))
        state.completed_steps.append("review")
        state.current_step = "completed"

        return state


# ============================================================================
# Workflow Builder
# ============================================================================

def create_multi_agent_workflow():
    """Create multi-agent LangGraph workflow"""
    if not LANGGRAPH_AVAILABLE:
        return None

    # Initialize agents
    research_agent = ResearchAgent()
    content_agent = ContentAgent()
    risk_agent = RiskAgent()
    review_agent = ReviewAgent()

    # Build workflow
    workflow = StateGraph(AgentState)

    # Add nodes
    workflow.add_node("research", research_agent)
    workflow.add_node("content", content_agent)
    workflow.add_node("risk", risk_agent)
    workflow.add_node("review", review_agent)

    # Set entry point
    workflow.set_entry_point("research")

    # Add edges (sequential workflow)
    workflow.add_edge("research", "content")
    workflow.add_edge("content", "risk")
    workflow.add_edge("risk", "review")
    workflow.add_edge("review", END)

    return workflow.compile()


# ============================================================================
# Campaign Orchestrator
# ============================================================================

class CampaignOrchestrator:
    """Orchestrates multi-agent phishing campaigns"""

    def __init__(self, llm_provider: str = "ollama", llm_model: str = "llama3.1:70b"):
        self.llm_provider = llm_provider
        self.llm_model = llm_model
        self.workflow = create_multi_agent_workflow()
        self.llm_client = None

        if OLLAMA_AVAILABLE and llm_provider == "ollama":
            self.llm_client = ollama.Client()

    def run_campaign(self, target_data: Dict[str, Any], campaign_id: str = "") -> Dict[str, Any]:
        """Run complete campaign generation workflow"""
        if not self.workflow:
            return {"error": "LangGraph not available", "fallback": True}

        initial_state = AgentState(
            messages=[],
            target_data=target_data,
            campaign_id=campaign_id or f"camp_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
        )

        # Run workflow
        final_state = self.workflow.invoke(initial_state)

        return {
            "success": True,
            "campaign_id": final_state.campaign_id,
            "output": final_state.final_output,
            "steps_completed": final_state.completed_steps,
            "errors": final_state.errors
        }

    def run_step(self, state: AgentState, step_name: str) -> AgentState:
        """Run specific workflow step"""
        if not self.workflow:
            return state

        # Get node
        node = self.workflow.get_node(step_name)
        if node:
            return node(state)

        return state


# ============================================================================
# FastAPI Integration
# ============================================================================

def setup_api_endpoints(app=None):
    """Setup FastAPI endpoints for agent orchestration"""
    if app is None:
        return None

    orchestrator = CampaignOrchestrator()

    @app.post("/api/v1/agents/run-campaign")
    async def run_campaign(target_data: dict, campaign_id: str = ""):
        """Run multi-agent campaign generation"""
        result = orchestrator.run_campaign(target_data, campaign_id)
        return result

    @app.post("/api/v1/agents/research")
    async def research_target(company: str, position: str):
        """Run research agent"""
        state = AgentState(target_data={"company": company, "position": position})
        result = ResearchAgent()(state)
        return {"research": result.research_results}

    @app.post("/api/v1/agents/generate-content")
    async def generate_content(template: str, target_data: dict):
        """Run content generation agent"""
        state = AgentState(target_data=target_data, generated_content={"template": template})
        result = ContentAgent()(state)
        return {"content": result.generated_content}

    @app.post("/api/v1/agents/assess-risk")
    async def assess_risk(target_data: dict):
        """Run risk assessment agent"""
        state = AgentState(target_data=target_data)
        result = RiskAgent()(state)
        return {"risk": result.risk_assessment}

    return True


# ============================================================================
# Main (for testing)
# ============================================================================

if __name__ == "__main__":
    if not LANGGRAPH_AVAILABLE:
        print("LangGraph not available - install with: pip install langgraph")
    else:
        # Test workflow
        orchestrator = CampaignOrchestrator()

        target = {
            "name": "John Smith",
            "email": "john.smith@acmecorp.com",
            "company": "Acme Corp",
            "position": "IT Manager",
            "interests": ["cybersecurity", "cloud computing"]
        }

        result = orchestrator.run_campaign(target, "test_campaign_001")
        print("\n=== Campaign Result ===")
        print(f"Campaign ID: {result.get('campaign_id')}")
        print(f"Success: {result.get('success')}")
        print(f"Steps: {result.get('steps_completed')}")

        if 'output' in result:
            output = result['output']
            print(f"\n=== Generated Email ===")
            print(f"Subject: {output['email']['subject']}")
            print(f"\nBody:\n{output['email']['body']}")
            print(f"\nSpam Score: {output['email']['spam_score']}")
            print(f"\nRisk Level: {output['risk']['risk_level']}")
            print(f"Success Probability: {output['risk']['success_probability']:.0%}")
