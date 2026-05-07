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
    doc.addImage("/logo.jpg", "JPG", 20, 10, 25, 25);
  } catch (e) {
    console.warn("Logo not found or could not be loaded");
  }

  // Header - Centered
  doc.setFont("times", "bold");
  doc.setFontSize(18);
  // Golden/Brownish color for company name
  doc.setTextColor(152, 131, 0); 
  doc.text("PT. Labersa Hutahaean", 105, 18, { align: "center" });
  
  doc.setFontSize(14);
  doc.setFont("times", "bold");
  // Dark Green color for branch
  doc.setTextColor(0, 100, 0);
  doc.text("HEAD OFFICE - WILAYAH TOBA", 105, 26, { align: "center" });

  doc.setDrawColor(0, 0, 0);
  doc.setLineWidth(0.5);
  doc.line(20, 38, 190, 38);

  // Title
  doc.setFont("times", "bold");
  doc.setFontSize(14);
  doc.setTextColor(0, 0, 0);
  doc.text("SURAT PERINTAH KERJA LEMBUR", 105, 50, { align: "center" });
  // Underline removed as requested

  // Instruction
  doc.setFont("times", "normal");
  doc.setFontSize(11);
  doc.text("Kepada saudara yang namanya tersebut di bawah ini diperintahkan kerja lembur:", 20, 65);

  // Overtime Details
  const detailsX = 20;
  const labelWidth = 40;
  const lineSpacing = 10; // Consistent vertical spacing
  let currentY = 75;

  // Row 1: Reason
  doc.text("Untuk keperluan / tugas", detailsX, currentY);
  doc.text(":", detailsX + labelWidth, currentY);
  const reasonText = overtime.reason || "-";
  const reasonLinesArr = doc.splitTextToSize(reasonText, 120);
  doc.text(reasonLinesArr, detailsX + labelWidth + 5, currentY);
  
  // Increment Y based on reason height
  currentY += Math.max(lineSpacing, reasonLinesArr.length * 5 + 3);

  // Row 2: Date
  doc.text("Pada hari / tanggal", detailsX, currentY);
  doc.text(":", detailsX + labelWidth, currentY);
  const formattedDate = new Date(overtime.date).toLocaleDateString("id-ID", { 
    weekday: 'long', 
    day: '2-digit', 
    month: 'long', 
    year: 'numeric' 
  });
  doc.text(formattedDate, detailsX + labelWidth + 5, currentY);
  
  currentY += lineSpacing;

  // Row 3: Time
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
      lineColor: [0, 0, 0],
      font: "times"
    },
    styles: { 
      fontSize: 11, 
      textColor: [0, 0, 0],
      lineWidth: 0.2,
      lineColor: [0, 0, 0],
      minCellHeight: 12,
      valign: "middle",
      font: "times"
    },
    columnStyles: {
      0: { cellWidth: 70 },
      1: { cellWidth: 60 },
      2: { cellWidth: 40 }
    }
  });

  // Signatures
  const footerY = (doc as any).lastAutoTable.finalY + 25;
  doc.setFont("times", "normal");
  doc.setFontSize(11);
  
  // Calculate center of each column (Content area is 170mm, 210mm wide)
  const col1X = 48; // Left center
  const col2X = 105; // Middle center
  const col3X = 162; // Right center

  // Signature Labels
  doc.text("Yang memberi perintah lembur,", col1X, footerY, { align: "center" });
  doc.text("Yang menerima perintah lembur,", col2X, footerY, { align: "center" });
  doc.text("Disetujui Oleh,", col3X, footerY, { align: "center" });

  const signLineY = footerY + 30;
  
  // Signature Lines
  doc.text("( ____________________ )", col1X, signLineY, { align: "center" });
  doc.text("( ____________________ )", col2X, signLineY, { align: "center" });
  doc.text("( ____________________ )", col3X, signLineY, { align: "center" });

  // Signature Sub-labels
  doc.setFontSize(9);
  doc.setFont("times", "bold");
  doc.text("Departement Head", col1X, signLineY + 6, { align: "center" });
  doc.text("Karyawan", col2X, signLineY + 6, { align: "center" });
  doc.text("Office Manager / HRM /", col3X, signLineY + 6, { align: "center" });
  doc.text("General Manager", col3X, signLineY + 11, { align: "center" });

  return doc.output("blob");
};



