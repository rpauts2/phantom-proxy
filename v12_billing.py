#!/usr/bin/env python3
"""
PhantomProxy v12.2 PRO+ — Billing & Invoice Module
Генерация профессиональных счетов для клиентов

© 2026 PhantomSec Labs. All rights reserved.
"""

import sqlite3
import json
from datetime import datetime, timedelta
from pathlib import Path
from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer
from reportlab.lib.units import inch

# === КОНФИГУРАЦИЯ ===
INVOICES_PATH = Path(__file__).parent / 'invoices'
TEMPLATES_PATH = Path(__file__).parent / 'templates'
DB_PATH = Path(__file__).parent / 'phantom.db'

# Создаём директории
INVOICES_PATH.mkdir(exist_ok=True)
TEMPLATES_PATH.mkdir(exist_ok=True)

# Company Info
COMPANY_INFO = {
    'name': 'PhantomSec Labs',
    'address': 'Москва, Россия',
    'email': 'info@phantomseclabs.com',
    'phone': '+7 (XXX) XXX-XX-XX',
    'website': 'https://phantomseclabs.com',
    'inn': 'XXXXXXXXXX',
    'kpp': 'XXXXXXXXX',
    'bank': 'ПАО Сбербанк',
    'bik': '044525225',
    'account': '40702810XXXXXXXXXXXXX'
}

class InvoiceGenerator:
    """Генерация профессиональных счетов"""
    
    def __init__(self, db_path=DB_PATH):
        self.db_path = db_path
        self.styles = getSampleStyleSheet()
        self._setup_styles()
    
    def _setup_styles(self):
        """Настройка стилей"""
        self.styles.add(ParagraphStyle(
            name='CompanyHeader',
            parent=self.styles['Heading1'],
            fontSize=18,
            textColor=colors.HexColor('#1E3A8A'),
            spaceAfter=6
        ))
        
        self.styles.add(ParagraphStyle(
            name='InvoiceTitle',
            parent=self.styles['Heading2'],
            fontSize=24,
            textColor=colors.HexColor('#EF4444'),
            spaceAfter=20,
            alignment=1  # Center
        ))
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def get_campaign_hours(self, campaign_id):
        """Расчёт часов кампании"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT created_at, stopped_at FROM campaigns WHERE id=?', (campaign_id,))
        campaign = c.fetchone()
        
        if campaign and campaign['created_at'] and campaign['stopped_at']:
            start = datetime.fromisoformat(campaign['created_at'])
            end = datetime.fromisoformat(campaign['stopped_at'])
            hours = (end - start).total_seconds() / 3600
        else:
            # Если не остановлена, считаем до текущего времени
            c.execute('SELECT created_at FROM campaigns WHERE id=?', (campaign_id,))
            campaign = c.fetchone()
            if campaign:
                start = datetime.fromisoformat(campaign['created_at'])
                hours = (datetime.now() - start).total_seconds() / 3600
            else:
                hours = 0
        
        conn.close()
        return round(hours, 2)
    
    def get_client_info(self, client_id):
        """Получение информации о клиенте"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT * FROM clients WHERE id=?', (client_id,))
        client = c.fetchone()
        
        conn.close()
        
        if client:
            return dict(client)
        return None
    
    def create_invoice(self, client_id, campaign_id, rate=500, due_days=30):
        """Создание счёта"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Получаем данные
        client = self.get_client_info(client_id)
        hours = self.get_campaign_hours(campaign_id)
        total = hours * rate
        
        # Получаем название кампании
        c.execute('SELECT name FROM campaigns WHERE id=?', (campaign_id,))
        campaign = c.fetchone()
        campaign_name = campaign['name'] if campaign else f'Campaign #{campaign_id}'
        
        # Генерируем номер
        c.execute('SELECT COUNT(*) FROM invoices')
        invoice_num = c.fetchone()[0] + 1
        invoice_number = f"INV-{invoice_num:04d}"
        
        # Даты
        issued_date = datetime.now()
        due_date = issued_date + timedelta(days=due_days)
        
        # Сохраняем в БД
        c.execute('''INSERT INTO invoices 
            (client_id, campaign_id, hours, rate, total, status, issued_date, due_date)
            VALUES (?, ?, ?, ?, ?, 'pending', ?, ?)''',
            (client_id, campaign_id, hours, rate, total, issued_date.isoformat(), due_date.isoformat()))
        
        invoice_id = c.lastrowid
        conn.commit()
        conn.close()
        
        # Генерируем PDF
        pdf_path = self.generate_pdf(invoice_id, invoice_number, client, campaign_name, hours, rate, total, issued_date, due_date)
        
        # Обновляем путь к PDF
        conn = self.get_db()
        c = conn.cursor()
        c.execute('UPDATE invoices SET pdf_path=? WHERE id=?', (str(pdf_path), invoice_id))
        conn.commit()
        conn.close()
        
        return {
            'invoice_id': invoice_id,
            'invoice_number': invoice_number,
            'pdf_path': pdf_path,
            'total': total,
            'hours': hours
        }
    
    def generate_pdf(self, invoice_id, invoice_number, client, campaign_name, hours, rate, total, issued_date, due_date):
        """Генерация PDF счёта"""
        output_path = INVOICES_PATH / f"{invoice_number}.pdf"
        
        doc = SimpleDocTemplate(
            str(output_path),
            pagesize=A4,
            rightMargin=0.75*inch,
            leftMargin=0.75*inch,
            topMargin=0.75*inch,
            bottomMargin=0.75*inch
        )
        
        elements = []
        
        # Header
        elements.append(Paragraph(f"{COMPANY_INFO['name']}", self.styles['CompanyHeader']))
        elements.append(Paragraph(f"{COMPANY_INFO['address']}", self.styles['Normal']))
        elements.append(Paragraph(f"{COMPANY_INFO['email']} | {COMPANY_INFO['phone']}", self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Invoice Title
        elements.append(Paragraph("INVOICE", self.styles['InvoiceTitle']))
        elements.append(Spacer(1, 0.2*inch))
        
        # Invoice Details
        invoice_data = [
            ['Invoice Number:', invoice_number],
            ['Issue Date:', issued_date.strftime('%Y-%m-%d')],
            ['Due Date:', due_date.strftime('%Y-%m-%d')],
            ['Status:', 'Pending']
        ]
        
        invoice_table = Table(invoice_data, colWidths=[2*inch, 2.5*inch])
        invoice_table.setStyle(TableStyle([
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (0, -1), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, -1), 10),
            ('BOTTOMPADDING', (0, 0), (-1, -1), 6),
        ]))
        elements.append(invoice_table)
        elements.append(Spacer(1, 0.3*inch))
        
        # Bill To
        elements.append(Paragraph("Bill To:", self.styles['Heading3']))
        if client:
            elements.append(Paragraph(f"{client.get('company_name', 'N/A')}", self.styles['Normal']))
            elements.append(Paragraph(f"{client.get('contact_email', 'N/A')}", self.styles['Normal']))
        else:
            elements.append(Paragraph("Client Information", self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Services Table
        elements.append(Paragraph("Services Rendered:", self.styles['Heading3']))
        elements.append(Spacer(1, 0.2*inch))
        
        services_data = [
            ['Description', 'Hours', 'Rate', 'Amount'],
            [f"Red Team Testing — {campaign_name}", str(hours), f"${rate}/hour", f"${total:,.2f}"],
            ['', '', 'Subtotal:', f"${total:,.2f}"],
            ['', '', 'Tax (0%):', '$0.00'],
            ['', '', 'Total:', f"${total:,.2f}"]
        ]
        
        services_table = Table(services_data, colWidths=[3*inch, 1*inch, 1.5*inch, 1.5*inch])
        services_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#1E3A8A')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'CENTER'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 11),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 12),
            ('BACKGROUND', (0, 1), (-1, 1), colors.HexColor('#f0f0f0')),
            ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
            ('FONTNAME', (0, 1), (-1, -1), 'Helvetica'),
            ('FONTSIZE', (0, 1), (-1, -1), 10),
            ('ALIGN', (2, 2), (-1, -1), 'RIGHT'),
            ('FONTNAME', (2, 4), (2, 4), 'Helvetica-Bold'),
        ]))
        elements.append(services_table)
        elements.append(Spacer(1, 0.5*inch))
        
        # Payment Info
        elements.append(Paragraph("Payment Information:", self.styles['Heading3']))
        payment_info = f"""
        Bank: {COMPANY_INFO['bank']}<br/>
        BIK: {COMPANY_INFO['bik']}<br/>
        Account: {COMPANY_INFO['account']}<br/>
        INN: {COMPANY_INFO['inn']}<br/>
        KPP: {COMPANY_INFO['kpp']}<br/>
        <br/>
        Please make payment within {due_days} days of invoice date.<br/>
        Thank you for your business!
        """
        elements.append(Paragraph(payment_info, self.styles['Normal']))
        elements.append(Spacer(1, 0.3*inch))
        
        # Footer
        elements.append(Paragraph(f"© 2026 {COMPANY_INFO['name']}. All rights reserved.", self.styles['Normal']))
        elements.append(Paragraph(f"{COMPANY_INFO['website']} | {COMPANY_INFO['email']}", self.styles['Normal']))
        
        # Build PDF
        doc.build(elements)
        
        return output_path
    
    def get_all_invoices(self):
        """Получение всех счетов"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT i.*, c.company_name as client_name 
                     FROM invoices i 
                     LEFT JOIN clients c ON i.client_id = c.id 
                     ORDER BY i.issued_date DESC''')
        
        invoices = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return invoices
    
    def mark_as_paid(self, invoice_id):
        """Отметка об оплате"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE invoices SET status=? WHERE id=?', ('paid', invoice_id))
        conn.commit()
        conn.close()
        
        return True

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.2 PRO+ — Billing Module")
    print("="*60)
    
    generator = InvoiceGenerator()
    print("✅ Invoice Generator initialized")
    print(f"📁 Invoices path: {INVOICES_PATH}")
    
    # Test
    print("\n📊 Getting all invoices...")
    invoices = generator.get_all_invoices()
    print(f"   Found {len(invoices)} invoices")
    
    print("\n💳 Sample invoice data:")
    if invoices:
        inv = invoices[0]
        print(f"   Invoice: {inv.get('invoice_number', 'N/A')}")
        print(f"   Client: {inv.get('client_name', 'N/A')}")
        print(f"   Total: ${inv.get('total', 0):,.2f}")
        print(f"   Status: {inv.get('status', 'N/A')}")
