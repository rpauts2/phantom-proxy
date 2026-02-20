#!/usr/bin/env python3
"""
PhantomProxy v12.2 PRO+ — Proposal & Contract Generator
Генерация шаблонов предложений и контрактов

© 2026 PhantomSec Labs. All rights reserved.
"""

import json
from datetime import datetime
from pathlib import Path
from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, PageBreak
from reportlab.lib.units import inch

# === КОНФИГУРАЦИЯ ===
PROPOSALS_PATH = Path(__file__).parent / 'proposals'
CONTRACTS_PATH = Path(__file__).parent / 'contracts'
DB_PATH = Path(__file__).parent / 'phantom.db'

# Создаём директории
PROPOSALS_PATH.mkdir(exist_ok=True)
CONTRACTS_PATH.mkdir(exist_ok=True)

# Company Info
COMPANY_INFO = {
    'name': 'PhantomSec Labs',
    'address': 'Москва, Россия',
    'email': 'info@phantomseclabs.com',
    'phone': '+7 (XXX) XXX-XX-XX',
    'website': 'https://phantomseclabs.com'
}

class ProposalGenerator:
    """Генерация коммерческих предложений"""
    
    def __init__(self):
        self.styles = getSampleStyleSheet()
        self._setup_styles()
    
    def _setup_styles(self):
        self.styles.add(ParagraphStyle(
            name='ProposalTitle',
            parent=self.styles['Heading1'],
            fontSize=24,
            textColor=colors.HexColor('#1E3A8A'),
            spaceAfter=30,
            alignment=1
        ))
    
    def generate_proposal(self, client_name, service_type='Red Team Assessment', duration='4 weeks', price=50000):
        """Генерация предложения"""
        output_path = PROPOSALS_PATH / f"Proposal_{client_name.replace(' ', '_')}_{datetime.now().strftime('%Y%m%d')}.pdf"
        
        doc = SimpleDocTemplate(str(output_path), pagesize=A4)
        elements = []
        
        # Header
        elements.append(Paragraph(f"{COMPANY_INFO['name']}", self.styles['Heading1']))
        elements.append(Paragraph(f"{COMPANY_INFO['address']}", self.styles['Normal']))
        elements.append(Paragraph(f"{COMPANY_INFO['email']} | {COMPANY_INFO['website']}", self.styles['Normal']))
        elements.append(Spacer(1, 0.5*inch))
        
        # Title
        elements.append(Paragraph("PROPOSAL", self.styles['ProposalTitle']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Client Info
        elements.append(Paragraph(f"Prepared for: {client_name}", self.styles['Heading3']))
        elements.append(Paragraph(f"Date: {datetime.now().strftime('%Y-%m-%d')}", self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Executive Summary
        elements.append(Paragraph("Executive Summary", self.styles['Heading2']))
        summary = f"""
        {COMPANY_INFO['name']} is pleased to present this proposal for {service_type} services.
        Our team of experienced security professionals will conduct a comprehensive assessment
        of your organization's security posture using advanced Red Team tactics and techniques.
        <br/><br/>
        This engagement will help identify vulnerabilities, test incident response capabilities,
        and provide actionable recommendations for improving your security posture.
        """
        elements.append(Paragraph(summary, self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Scope of Work
        elements.append(Paragraph("Scope of Work", self.styles['Heading2']))
        
        scope_data = [
            ['Phase', 'Description', 'Duration'],
            ['Reconnaissance', 'Open-source intelligence gathering, social media analysis', '3 days'],
            ['Initial Access', 'Phishing campaigns, social engineering', '5 days'],
            ['Persistence', 'Establishing foothold, lateral movement', '1 week'],
            ['Exfiltration', 'Data exfiltration simulation', '1 week'],
            ['Reporting', 'Detailed findings and recommendations', '1 week']
        ]
        
        scope_table = Table(scope_data, colWidths=[1.5*inch, 3*inch, 1.5*inch])
        scope_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1E3A8A')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
            ('FONTSIZE', (0, 0), (-1, -1), 9),
        ]))
        elements.append(scope_table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Timeline
        elements.append(Paragraph("Timeline", self.styles['Heading2']))
        elements.append(Paragraph(f"Total Duration: {duration}", self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Pricing
        elements.append(Paragraph("Investment", self.styles['Heading2']))
        
        pricing_data = [
            ['Service', 'Price'],
            [service_type, f"${price:,.2f}'],
            ['Additional Days (if needed)', '$1,500/day'],
            ['Retesting', '$5,000'],
            ['', ''],
            ['Total', f"${price:,.2f}']
        ]
        
        pricing_table = Table(pricing_data, colWidths=[4*inch, 2*inch])
        pricing_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1E3A8A')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
            ('FONTSIZE', (0, 0), (-1, -1), 10),
            ('BACKGROUND', (0, 5), (-1, 5), colors.HexColor('#f0f0f0')),
            ('FONTNAME', (0, 5), (-1, 5), 'Helvetica-Bold'),
        ]))
        elements.append(pricing_table)
        elements.append(Spacer(1, 0.5*inch))
        
        # Terms
        elements.append(Paragraph("Terms and Conditions", self.styles['Heading2']))
        terms = """
        • 50% deposit required to begin engagement<br/>
        • Balance due upon completion<br/>
        • All payments due within 30 days of invoice date<br/>
        • Client must provide written authorization (RoE) before engagement begins<br/>
        • Engagement subject to our standard terms and conditions<br/>
        • Confidentiality agreement included<br/>
        """
        elements.append(Paragraph(terms, self.styles['Normal']))
        elements.append(Spacer(1, 0.5*inch))
        
        # Contact
        elements.append(Paragraph("Next Steps", self.styles['Heading2']))
        contact = f"""
        To proceed with this engagement, please:<br/>
        1. Sign and return this proposal<br/>
        2. Provide signed Rules of Engagement (RoE)<br/>
        3. Submit deposit payment<br/>
        <br/>
        Contact us at {COMPANY_INFO['email']} or {COMPANY_INFO['phone']} with any questions.<br/>
        <br/>
        We look forward to working with you!<br/>
        <br/>
        Sincerely,<br/>
        The {COMPANY_INFO['name']} Team
        """
        elements.append(Paragraph(contact, self.styles['Normal']))
        
        # Build PDF
        doc.build(elements)
        
        return output_path

class ContractGenerator:
    """Генерация контрактов и RoE"""
    
    def __init__(self):
        self.styles = getSampleStyleSheet()
    
    def generate_roe(self, client_name, campaign_name, start_date, end_date, authorized_ips):
        """Генерация Rules of Engagement"""
        output_path = CONTRACTS_PATH / f"RoE_{client_name.replace(' ', '_')}_{datetime.now().strftime('%Y%m%d')}.pdf"
        
        doc = SimpleDocTemplate(str(output_path), pagesize=A4)
        elements = []
        
        # Header
        elements.append(Paragraph(f"{COMPANY_INFO['name']}", self.styles['Heading1']))
        elements.append(Paragraph("RULES OF ENGAGEMENT (RoE)", self.styles['Heading2']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Project Info
        elements.append(Paragraph("Project Information", self.styles['Heading3']))
        info_data = [
            ['Client:', client_name],
            ['Campaign:', campaign_name],
            ['Start Date:', start_date],
            ['End Date:', end_date],
            ['Authorized IPs:', ', '.join(authorized_ips)]
        ]
        info_table = Table(info_data, colWidths=[2*inch, 4*inch])
        info_table.setStyle(TableStyle([
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (0, -1), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, -1), 10),
            ('BOTTOMPADDING', (0, 0), (-1, -1), 6),
        ]))
        elements.append(info_table)
        elements.append(Spacer(1, 0.5*inch))
        
        # Authorized Activities
        elements.append(Paragraph("Authorized Activities", self.styles['Heading3']))
        activities = """
        ✓ Phishing simulations (email, SMS, voice)<br/>
        ✓ Social engineering engagements<br/>
        ✓ Physical security testing (if authorized)<br/>
        ✓ Network penetration testing<br/>
        ✓ Web application testing<br/>
        ✓ Wireless network assessment<br/>
        """
        elements.append(Paragraph(activities, self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Prohibited Activities
        elements.append(Paragraph("Prohibited Activities", self.styles['Heading3']))
        prohibited = """
        ✗ Denial of Service (DoS) attacks<br/>
        ✗ Ransomware deployment<br/>
        ✗ Data destruction or modification<br/>
        ✗ Attacks outside authorized scope<br/>
        ✗ Social engineering of minors<br/>
        ✗ Physical harm or threats<br/>
        """
        elements.append(Paragraph(prohibited, self.styles['Normal']))
        elements.append(Spacer(1, 0.5*inch))
        
        # Emergency Contacts
        elements.append(Paragraph("Emergency Contacts", self.styles['Heading3']))
        contacts = f"""
        {COMPANY_INFO['name']}: {COMPANY_INFO['phone']}<br/>
        Client Primary Contact: [TO BE PROVIDED]<br/>
        Client Secondary Contact: [TO BE PROVIDED]<br/>
        """
        elements.append(Paragraph(contacts, self.styles['Normal']))
        
        # Signatures
        elements.append(PageBreak())
        elements.append(Paragraph("Signatures", self.styles['Heading2']))
        elements.append(Spacer(1, 1*inch))
        
        signature_data = [
            ['_________________________', '_________________________'],
            [f'{COMPANY_INFO["name"]}', client_name],
            ['Authorized Signature', 'Client Signature'],
            ['Date: _______________', 'Date: _______________']
        ]
        
        sig_table = Table(signature_data, colWidths=[3*inch, 3*inch])
        sig_table.setStyle(TableStyle([
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, -1), 'Helvetica'),
            ('FONTSIZE', (0, 0), (-1, -1), 11),
            ('BOTTOMPADDING', (0, 0), (-1, -1), 12),
        ]))
        elements.append(sig_table)
        
        # Build PDF
        doc.build(elements)
        
        return output_path
    
    def generate_nda(self, client_name):
        """Генерация NDA"""
        output_path = CONTRACTS_PATH / f"NDA_{client_name.replace(' ', '_')}_{datetime.now().strftime('%Y%m%d')}.pdf"
        
        doc = SimpleDocTemplate(str(output_path), pagesize=A4)
        elements = []
        
        elements.append(Paragraph(f"{COMPANY_INFO['name']}", self.styles['Heading1']))
        elements.append(Paragraph("NON-DISCLOSURE AGREEMENT", self.styles['Heading2']))
        elements.append(Spacer(1, 0.5*inch))
        
        nda_text = f"""
        This Non-Disclosure Agreement ("Agreement") is entered into on {datetime.now().strftime('%Y-%m-%d')}
        by and between {COMPANY_INFO['name']} ("Disclosing Party") and {client_name} ("Receiving Party").
        <br/><br/>
        <b>1. Confidential Information</b><br/>
        For purposes of this Agreement, "Confidential Information" shall include all information
        disclosed during Red Team engagement, including but not limited to:<br/>
        • Security vulnerabilities and findings<br/>
        • Technical data and methodologies<br/>
        • Client employee information<br/>
        • Security infrastructure details<br/>
        <br/><br/>
        <b>2. Obligations</b><br/>
        Receiving Party agrees to:<br/>
        • Hold Confidential Information in strict confidence<br/>
        • Not disclose to third parties without written consent<br/>
        • Use information solely for authorized security testing<br/>
        • Implement reasonable security measures<br/>
        <br/><br/>
        <b>3. Term</b><br/>
        This Agreement shall remain in effect for a period of five (5) years from the date of execution.<br/>
        <br/><br/>
        <b>4. Governing Law</b><br/>
        This Agreement shall be governed by the laws of the Russian Federation.<br/>
        """
        
        elements.append(Paragraph(nda_text, self.styles['Normal']))
        
        # Build PDF
        doc.build(elements)
        
        return output_path

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.2 PRO+ — Proposal & Contract Generator")
    print("="*70)
    
    # Test Proposal
    print("\n📄 Generating Proposal...")
    proposal_gen = ProposalGenerator()
    proposal_path = proposal_gen.generate_proposal(
        client_name='Test Client',
        service_type='Red Team Assessment',
        duration='4 weeks',
        price=50000
    )
    print(f"✅ Proposal generated: {proposal_path}")
    
    # Test RoE
    print("\n📋 Generating RoE...")
    contract_gen = ContractGenerator()
    roe_path = contract_gen.generate_roe(
        client_name='Test Client',
        campaign_name='Q1 Phishing Campaign',
        start_date='2026-03-01',
        end_date='2026-03-31',
        authorized_ips=['192.168.1.100', '10.0.0.50']
    )
    print(f"✅ RoE generated: {roe_path}")
    
    # Test NDA
    print("\n🔒 Generating NDA...")
    nda_path = contract_gen.generate_nda(client_name='Test Client')
    print(f"✅ NDA generated: {nda_path}")
    
    print("\n✅ All templates generated successfully!")
