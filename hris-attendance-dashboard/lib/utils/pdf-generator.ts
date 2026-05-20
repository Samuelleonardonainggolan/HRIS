// lib/utils/pdf-generator.ts
// NOTE: jsPDF is browser-only. Never import/call this from server components.

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

export const generateSPKLPDF = async (
  overtime: OvertimeData,
  employee: EmployeeData
): Promise<Blob> => {
  // Dynamic import to avoid SSR issues
  const { jsPDF } = await import("jspdf");
  const { default: autoTable } = await import("jspdf-autotable");

  const doc = new jsPDF({ orientation: "portrait", unit: "mm", format: "a4" });

  // Logo
  try {
    doc.addImage("/logo.jpg", "JPG", 20, 10, 25, 25);
  } catch {
    // skip if logo not found
  }

  // Header
  doc.setFont("helvetica", "bold");
  doc.setFontSize(18);
  doc.setTextColor(152, 131, 0);
  doc.text("PT. Labersa Hutahaean", 105, 18, { align: "center" });

  doc.setFontSize(14);
  doc.setTextColor(0, 100, 0);
  doc.text("HEAD OFFICE - WILAYAH TOBA", 105, 26, { align: "center" });

  doc.setDrawColor(0, 0, 0);
  doc.setLineWidth(0.5);
  doc.line(20, 38, 190, 38);

  // Title
  doc.setFont("helvetica", "bold");
  doc.setFontSize(14);
  doc.setTextColor(0, 0, 0);
  doc.text("SURAT PERINTAH KERJA LEMBUR", 105, 50, { align: "center" });

  doc.setFont("helvetica", "normal");
  doc.setFontSize(11);
  doc.text(
    "Kepada saudara yang namanya tersebut di bawah ini diperintahkan kerja lembur:",
    20,
    65
  );

  const detailsX = 20;
  const labelWidth = 45;
  const lineSpacing = 10;
  let currentY = 75;

  const formattedDate = new Date(overtime.date).toLocaleDateString("id-ID", {
    weekday: "long",
    day: "2-digit",
    month: "long",
    year: "numeric",
  });

  const rows: [string, string][] = [
    ["Untuk keperluan / tugas", overtime.reason || "-"],
    ["Pada hari / tanggal", formattedDate],
    ["Dimulai jam", (overtime.start_time || "").slice(0, 5) + " WIB"],
    ["Selesai jam", (overtime.end_time || "").slice(0, 5) + " WIB"],
  ];

  for (const [label, value] of rows) {
    doc.text(label, detailsX, currentY);
    doc.text(":", detailsX + labelWidth, currentY);
    const lines = doc.splitTextToSize(value, 115);
    doc.text(lines, detailsX + labelWidth + 5, currentY);
    currentY += Math.max(lineSpacing, lines.length * 6);
  }

  currentY += 5;

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
    },
    styles: {
      fontSize: 11,
      textColor: [0, 0, 0],
      lineWidth: 0.2,
      lineColor: [0, 0, 0],
      minCellHeight: 12,
      valign: "middle",
    },
    columnStyles: {
      0: { cellWidth: 70 },
      1: { cellWidth: 60 },
      2: { cellWidth: 40 },
    },
  });

  const footerY = (doc as any).lastAutoTable.finalY + 25;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(11);

  const col1X = 48;
  const col2X = 105;
  const col3X = 162;

  doc.text("Yang memberi perintah lembur,", col1X, footerY, { align: "center" });
  doc.text("Yang menerima perintah lembur,", col2X, footerY, { align: "center" });
  doc.text("Disetujui Oleh,", col3X, footerY, { align: "center" });

  const signLineY = footerY + 30;
  doc.text("( ____________________ )", col1X, signLineY, { align: "center" });
  doc.text("( ____________________ )", col2X, signLineY, { align: "center" });
  doc.text("( ____________________ )", col3X, signLineY, { align: "center" });

  doc.setFontSize(9);
  doc.setFont("helvetica", "bold");
  doc.text("Departement Head", col1X, signLineY + 6, { align: "center" });
  doc.text("Karyawan", col2X, signLineY + 6, { align: "center" });
  doc.text("Office Manager / HRM /", col3X, signLineY + 6, { align: "center" });
  doc.text("General Manager", col3X, signLineY + 11, { align: "center" });

  return doc.output("blob");
};

interface AttendanceReportData {
  title: string;
  period: string;
  headers: string[];
  body: any[][];
}

export const generateAttendanceReportPDF = async (
  report: AttendanceReportData
): Promise<Blob> => {
  // Dynamic import to avoid SSR issues
  const { jsPDF } = await import("jspdf");
  const { default: autoTable } = await import("jspdf-autotable");

  const doc = new jsPDF({ orientation: "landscape", unit: "mm", format: "a4" });

  // Logo
  try {
    doc.addImage("/logo.jpg", "JPG", 20, 10, 20, 20);
  } catch {
    // skip if logo not found
  }

  // Company Header
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.setTextColor(152, 131, 0);
  doc.text("PT. Labersa Hutahaean", 148, 15, { align: "center" });

  doc.setFontSize(12);
  doc.setTextColor(0, 100, 0);
  doc.text("HEAD OFFICE - WILAYAH TOBA", 148, 22, { align: "center" });

  doc.setDrawColor(0, 0, 0);
  doc.setLineWidth(0.5);
  doc.line(20, 32, 277, 32);

  // Report Title
  doc.setFont("helvetica", "bold");
  doc.setFontSize(14);
  doc.setTextColor(0, 0, 0);
  doc.text(report.title.toUpperCase(), 148, 42, { align: "center" });

  doc.setFontSize(11);
  doc.setFont("helvetica", "normal");
  doc.text(`Periode: ${report.period}`, 148, 49, { align: "center" });

  // Table
  autoTable(doc, {
    startY: 56,
    margin: { left: 15, right: 15 },
    head: [report.headers],
    body: report.body,
    theme: "grid",
    headStyles: {
      fillColor: [37, 99, 235],   // Blue-600
      textColor: [255, 255, 255],
      fontStyle: "bold",
      halign: "center",
      lineWidth: 0.1,
      lineColor: [0, 0, 0],
      fontSize: 9,
    },
    alternateRowStyles: {
      fillColor: [239, 246, 255], // Blue-50
    },
    styles: {
      fontSize: 8,
      textColor: [0, 0, 0],
      lineWidth: 0.1,
      lineColor: [200, 200, 200],
      valign: "middle",
      overflow: "linebreak",
    },
    columnStyles: {
      0: { cellWidth: 25 },  // Tanggal
      1: { cellWidth: 45 },  // Karyawan
      2: { cellWidth: 20 },  // NIK
      3: { cellWidth: 22 },  // Jenis
      4: { cellWidth: "auto" }, // Detail
      5: { cellWidth: 22 },  // Status
    },
  });

  // Footer
  const finalY = (doc as any).lastAutoTable?.finalY ?? 56;
  doc.setFontSize(8);
  doc.setFont("helvetica", "normal");
  doc.setTextColor(100, 100, 100);
  doc.text(
    `Dicetak pada: ${new Date().toLocaleString("id-ID")}`,
    15,
    Math.min(finalY + 8, 200)
  );

  return doc.output("blob");
};
