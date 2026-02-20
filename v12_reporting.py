#!/usr/bin/env python3
"""
PhantomProxy v12.0 - RED TEAM PROFESSIONAL EDITION
Модуль: Professional Reporting Engine (PDF + Evidence Collection)

Для легального использования в рамках Red Team engagements
Только для аккредитованных организаций с письменными разрешениями
"""

import os
import json
import sqlite3
import hashlib
from datetime import datetime
from pathlib import Path
from reportlab.lib import colors
from reportlab.lib.pagesizes import A4, landscape
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, Image, PageBreak
from reportlab.lib.enums import TA_CENTER, TA_LEFT
import io

# === КОНФИГУРАЦИЯ ===
REPORTS_PATH = Path(__file__).parent / 'reports'
EVIDENCE_PATH = Path(__file__).parent / 'evidence'
TEMPLATES_PATH = Path(__file__).parent / 'templates'
DB_PATH = Path(__file__).parent / 'phantom.db'

# Создаём директории
REPORTS_PATH.mkdir(exist_ok=True)
EVIDENCE_PATH.mkdir(exist_ok=True)

class ReportGenerator:
    """Генерация профессиональных PDF отчётов для клиентов"""
    
    def __init__(self, db_path=DB_PATH):
        self.db_path = db_path
        self.styles = getSampleStyleSheet()
        self._setup_styles()
    
    def _setup_styles(self):
        """Настройка стилей для отчёта"""
        self.styles.add(ParagraphStyle(
            name='CustomTitle',
            parent=self.styles['Heading1'],
            fontSize=24,
            textColor=colors.HexColor('#1a1a2e'),
            spaceAfter=30,
            alignment=TA_CENTER
        ))
        
        self.styles.add(ParagraphStyle(
            name='CompanyInfo',
            parent=self.styles['Normal'],
            fontSize=10,
            textColor=colors.HexColor('#666666'),
            spaceAfter=6
        ))
        
        self.styles.add(ParagraphStyle(
            name='SectionHeader',
            parent=self.styles['Heading2'],
            fontSize=16,
            textColor=colors.HexColor('#16213e'),
            spaceAfter=12,
            spaceBefore=12
        ))
    
    def get_campaign_stats(self, campaign_id=None):
        """Получение статистики кампании"""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        c = conn.cursor()
        
        if campaign_id:
            c.execute('SELECT * FROM campaigns WHERE id=?', (campaign_id,))
            campaign = dict(c.fetchone())
            c.execute('SELECT * FROM sessions WHERE campaign_id=?', (campaign_id,))
        else:
            campaign = None
            c.execute('SELECT * FROM sessions')
        
        sessions = [dict(row) for row in c.fetchall()]
        
        # Статистика
        total = len(sessions)
        excellent = sum(1 for s in sessions if s.get('classification') == 'EXCELLENT')
        good = sum(1 for s in sessions if s.get('classification') == 'GOOD')
        average = sum(1 for s in sessions if s.get('classification') == 'AVERAGE')
        low = sum(1 for s in sessions if s.get('classification') == 'LOW')
        
        # По сервисам
        services = {}
        for s in sessions:
            service = s.get('service', 'Unknown')
            services[service] = services.get(service, 0) + 1
        
        # По времени
        if sessions:
            first = min(s['created_at'] for s in sessions)
            last = max(s['created_at'] for s in sessions)
        else:
            first = last = 'N/A'
        
        conn.close()
        
        return {
            'campaign': campaign,
            'total_sessions': total,
            'quality_breakdown': {
                'EXCELLENT': excellent,
                'GOOD': good,
                'AVERAGE': average,
                'LOW': low
            },
            'services': services,
            'date_range': {'start': first, 'end': last},
            'sessions': sessions
        }
    
    def generate_pdf_report(self, campaign_id=None, client_name='Client', output_path=None):
        """Генерация PDF отчёта"""
        stats = self.get_campaign_stats(campaign_id)
        
        if output_path is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            output_path = REPORTS_PATH / f"Report_{client_name}_{timestamp}.pdf"
        
        doc = SimpleDocTemplate(
            str(output_path),
            pagesize=landscape(A4),
            rightMargin=0.5*inch,
            leftMargin=0.5*inch,
            topMargin=0.5*inch,
            bottomMargin=0.5*inch
        )
        
        elements = []
        
        # Title Page
        elements.append(Paragraph("PhantomProxy Red Team Report", self.styles['CustomTitle']))
        elements.append(Spacer(1, 0.3*inch))
        
        elements.append(Paragraph(f"Client: {client_name}", self.styles['Heading3']))
        elements.append(Paragraph(f"Report Date: {datetime.now().strftime('%Y-%m-%d %H:%M')}", self.styles['CompanyInfo']))
        if stats['campaign']:
            elements.append(Paragraph(f"Campaign: {stats['campaign'].get('name', 'N/A')}", self.styles['CompanyInfo']))
        elements.append(Spacer(1, 0.5*inch))
        
        # Executive Summary
        elements.append(Paragraph("Executive Summary", self.styles['SectionHeader']))
        summary_text = f"""
        This report presents the results of the phishing simulation campaign conducted for {client_name}. 
        The campaign was executed under authorized Rules of Engagement (RoE) for the purpose of 
        evaluating organizational security posture and employee awareness.
        <br/><br/>
        <b>Total Sessions Captured:</b> {stats['total_sessions']}<br/>
        <b>Campaign Period:</b> {stats['date_range']['start']} to {stats['date_range']['end']}<br/>
        <br/><br/>
        <b>Key Findings:</b><br/>
        • {stats['quality_breakdown']['EXCELLENT'] + stats['quality_breakdown']['GOOD']} high-quality credentials captured<br/>
        • {len(stats['services'])} different services targeted<br/>
        • Campaign effectiveness: {self._calculate_effectiveness(stats)}%
        """
        elements.append(Paragraph(summary_text, self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Statistics Table
        elements.append(Paragraph("Session Statistics", self.styles['SectionHeader']))
        
        data = [['Metric', 'Value']]
        data.append(['Total Sessions', str(stats['total_sessions'])])
        data.append(['Excellent Quality', str(stats['quality_breakdown']['EXCELLENT'])])
        data.append(['Good Quality', str(stats['quality_breakdown']['GOOD'])])
        data.append(['Average Quality', str(stats['quality_breakdown']['AVERAGE'])])
        data.append(['Low Quality', str(stats['quality_breakdown']['LOW'])])
        data.append(['Services Targeted', str(len(stats['services']))])
        
        table = Table(data, colWidths=[3*inch, 2*inch])
        table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#16213e')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 12),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 12),
            ('BACKGROUND', (0, 1), (-1, -1), colors.HexColor('#f0f0f0')),
            ('TEXTCOLOR', (0, 1), (-1, -1), colors.black),
            ('FONTNAME', (0, 1), (-1, -1), 'Helvetica'),
            ('FONTSIZE', (0, 1), (-1, -1), 10),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
        ]))
        elements.append(table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Services Breakdown
        elements.append(Paragraph("Services Targeted", self.styles['SectionHeader']))
        
        service_data = [['Service', 'Sessions', 'Percentage']]
        for service, count in sorted(stats['services'].items(), key=lambda x: x[1], reverse=True):
            percentage = round(count / max(1, stats['total_sessions']) * 100, 1)
            service_data.append([service, str(count), f'{percentage}%'])
        
        service_table = Table(service_data, colWidths=[2.5*inch, 1.5*inch, 1.5*inch])
        service_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#16213e')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
        ]))
        elements.append(service_table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Session Details
        elements.append(Paragraph("Session Details (Last 50)", self.styles['SectionHeader']))
        
        session_data = [['ID', 'Email', 'Service', 'Quality', 'Timestamp']]
        for s in stats['sessions'][:50]:
            session_data.append([
                str(s['id']),
                s.get('email', 'N/A')[:30],
                s.get('service', 'N/A')[:20],
                s.get('classification', 'N/A'),
                s.get('created_at', 'N/A')[:16]
            ])
        
        session_table = Table(session_data, colWidths=[0.8*inch, 2.5*inch, 1.5*inch, 1*inch, 1.5*inch])
        session_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#16213e')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, -1), 8),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
        ]))
        elements.append(session_table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Recommendations
        elements.append(Paragraph("Recommendations", self.styles['SectionHeader']))
        recommendations = """
        <b>Immediate Actions:</b><br/>
        1. Reset credentials for all compromised accounts<br/>
        2. Enable MFA for affected users<br/>
        3. Review SIEM logs for suspicious activity<br/>
        <br/>
        <b>Long-term Improvements:</b><br/>
        1. Implement security awareness training<br/>
        2. Deploy advanced email filtering<br/>
        3. Conduct regular phishing simulations<br/>
        4. Establish incident response procedures<br/>
        """
        elements.append(Paragraph(recommendations, self.styles['Normal']))
        
        # Footer
        elements.append(PageBreak())
        elements.append(Paragraph("Report Classification: CONFIDENTIAL", self.styles['CompanyInfo']))
        elements.append(Paragraph("This report contains sensitive information and should be handled accordingly.", self.styles['CompanyInfo']))
        elements.append(Paragraph(f"Generated by PhantomProxy v12.0 RED TEAM PROFESSIONAL", self.styles['CompanyInfo']))
        elements.append(Paragraph(f"Report generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}", self.styles['CompanyInfo']))
        
        # Build PDF
        doc.build(elements)
        
        return output_path
    
    def _calculate_effectiveness(self, stats):
        """Расчёт эффективности кампании"""
        if stats['total_sessions'] == 0:
            return 0
        high_quality = stats['quality_breakdown']['EXCELLENT'] + stats['quality_breakdown']['GOOD']
        return round(high_quality / stats['total_sessions'] * 100, 1)
    
    def collect_evidence(self, session_id):
        """Сбор доказательств для сессии"""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        c = conn.cursor()
        c.execute('SELECT * FROM sessions WHERE id=?', (session_id,))
        session = dict(c.fetchone())
        conn.close()
        
        evidence = {
            'session_id': session_id,
            'timestamp': datetime.now().isoformat(),
            'data': session,
            'hash': hashlib.sha256(json.dumps(session, sort_keys=True).encode()).hexdigest()
        }
        
        # Сохранение
        evidence_file = EVIDENCE_PATH / f"evidence_{session_id}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(evidence_file, 'w') as f:
            json.dump(evidence, f, indent=2)
        
        return evidence_file

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.0 - Reporting Engine Module")
    print("="*60)
    
    generator = ReportGenerator()
    print("✅ Reporting Engine initialized")
    print(f"📁 Reports path: {REPORTS_PATH}")
    print(f"📁 Evidence path: {EVIDENCE_PATH}")
    
    # Test report generation
    print("\n📊 Generating test report...")
    try:
        report_path = generator.generate_pdf_report(client_name='Test Client')
        print(f"✅ Report generated: {report_path}")
    except Exception as e:
        print(f"⚠️ Report generation requires reportlab: pip install reportlab")
        print(f"   Error: {e}")
