import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";

interface OvertimeData {
  id: string;
  date: string;
  start_time: string;
  end_time: string;
  reason: string;
  department_name: string;
}

interface EmployeeData {
  full_name: string;
  payroll_number: string;
  position_name: string;
}

export const generateSPKLPDF = async (overtime: OvertimeData, employee: EmployeeData): Promise<Blob> => {
  const doc = new jsPDF({
    orientation: "portrait",
    unit: "mm",
    format: "a4",
  });

  // Helper for Logo (Try to load it if exists)
  try {
    // Note: In a real environment, you might want to use a base64 string for the logo 
    // to ensure it's always available and loads instantly.
    doc.addImage("/logo.png", "PNG", 20, 10, 25, 25);
  } catch (e) {
    console.warn("Logo not found or could not be loaded");
  }

  // Header
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.setTextColor(0, 0, 0);
  doc.text("PT. LABERSA HUTAHAEAN", 50, 18);
  
  doc.setFontSize(11);
  doc.setFont("helvetica", "normal");
  doc.text("HEAD OFFICE - WILAYAH TOBA", 50, 24);

  doc.setLineWidth(0.5);
  doc.line(20, 38, 190, 38);

  // Title
  doc.setFont("helvetica", "bold");
  doc.setFontSize(14);
  doc.text("SURAT PERINTAH KERJA LEMBUR", 105, 50, { align: "center" });
  doc.line(65, 51, 145, 51); // Underline title

  // Instruction
  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.text("Kepada saudara yang namanya tersebut di bawah ini diperintahkan kerja lembur:", 20, 65);

  // Overtime Details
  const detailsX = 20;
  const labelWidth = 40;
  let currentY = 75;

  doc.text("Untuk keperluan / tugas", detailsX, currentY);
  doc.text(":", detailsX + labelWidth, currentY);
  doc.text(overtime.reason, detailsX + labelWidth + 5, currentY, { maxWidth: 120 });
  
  // Calculate height of reason text to adjust next Y
  const reasonLines = doc.splitTextToSize(overtime.reason, 120).length;
  currentY += Math.max(6, reasonLines * 5);

  doc.text("Pada hari / tanggal", detailsX, currentY);
  doc.text(":", detailsX + labelWidth, currentY);
  const formattedDate = new Date(overtime.date).toLocaleDateString("id-ID", { 
    weekday: 'long', 
    day: '2-digit', 
    month: 'long', 
    year: 'numeric' 
  });
  doc.text(formattedDate, detailsX + labelWidth + 5, currentY);
  currentY += 8;

  doc.text("Dimulai jam", detailsX, currentY);
  doc.text(":", detailsX + labelWidth, currentY);
  doc.text(overtime.start_time.slice(0, 5) + " WIB", detailsX + labelWidth + 5, currentY);
  currentY += 15;

  // Employee Table
  autoTable(doc, {
    startY: currentY,
    margin: { left: 20, right: 20 },
    head: [["Nama", "Jabatan", "Tanda Tangan"]],
    body: [[employee.full_name, employee.position_name, ""]],
    theme: "grid",
    headStyles: { 
      fillColor: [255, 255, 255], 
      textColor: [0, 0, 0], 
      fontStyle: "bold", 
      halign: "center",
      lineWidth: 0.2,
      lineColor: [0, 0, 0]
    },
    styles: { 
      fontSize: 10, 
      textColor: [0, 0, 0],
      lineWidth: 0.2,
      lineColor: [0, 0, 0],
      minCellHeight: 12,
      valign: "middle"
    },
    columnStyles: {
      0: { cellWidth: 70 },
      1: { cellWidth: 60 },
      2: { cellWidth: 40 }
    }
  });

  // Signatures
  const footerY = (doc as any).lastAutoTable.finalY + 25;
  doc.setFontSize(9);
  
  // Signature Labels
  doc.text("Yang memberi perintah lembur,", 20, footerY);
  doc.text("Yang menerima perintah lembur,", 85, footerY);
  doc.text("Disetujui Oleh,", 150, footerY);

  const signLineY = footerY + 25;
  
  // Parentheses for names
  doc.text("( ____________________ )", 20, signLineY);
  doc.text("( ____________________ )", 85, signLineY);
  doc.text("( ____________________ )", 150, signLineY);

  // Signature Sub-labels
  doc.setFontSize(8);
  doc.setFont("helvetica", "bold");
  doc.text("Departement Head", 35, signLineY + 5, { align: "center" });
  doc.text("Karyawan", 100, signLineY + 5, { align: "center" });
  doc.text("Office Manager / HRM /", 165, signLineY + 5, { align: "center" });
  doc.text("General Manager", 165, signLineY + 9, { align: "center" });

  return doc.output("blob");
};

