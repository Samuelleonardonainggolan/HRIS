"use client";

import { useState } from "react";
import { Upload, FileSpreadsheet, Download, FileCheck } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface ImportExcelModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

type Step = 1 | 2 | 3;

export function ImportExcelModal({ open, onOpenChange }: ImportExcelModalProps) {
  const [currentStep, setCurrentStep] = useState<Step>(1);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const handleFileSelect = (file: File) => {
    // Validate file type
    const validTypes = [
      "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      ".xlsx",
      ".xls",
    ];
    
    const isValid = validTypes.some(type => 
      file.type === type || file.name.toLowerCase().endsWith(type)
    );

    if (!isValid) {
      alert("Format file tidak valid. Hanya file Excel (.xlsx, .xls) yang diperbolehkan.");
      return;
    }

    // Validate file size (max 5MB)
    const maxSize = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSize) {
      alert("Ukuran file terlalu besar. Maksimal 5MB.");
      return;
    }

    setSelectedFile(file);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  };

  const handleDownloadTemplate = () => {
    // TODO: Implement download template
    console.log("Download template");
  };

  const handleNext = () => {
    if (currentStep === 1 && !selectedFile) {
      alert("Silakan pilih file terlebih dahulu");
      return;
    }
    if (currentStep < 3) {
      setCurrentStep((currentStep + 1) as Step);
    }
  };

  const handleCancel = () => {
    setCurrentStep(1);
    setSelectedFile(null);
    onOpenChange(false);
  };

  const handleComplete = () => {
    // TODO: Implement import logic
    console.log("Import file:", selectedFile);
    setCurrentStep(1);
    setSelectedFile(null);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Impor Data Pegawai</DialogTitle>
        </DialogHeader>

        {/* Stepper */}
        <div className="flex items-center justify-center gap-4 py-4">
          {/* Step 1 */}
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold",
                currentStep >= 1
                  ? "bg-blue-600 text-white"
                  : "bg-gray-200 text-gray-600"
              )}
            >
              1
            </div>
            <span
              className={cn(
                "text-sm font-medium",
                currentStep >= 1 ? "text-blue-600" : "text-gray-500"
              )}
            >
              Unggah File
            </span>
          </div>

          {/* Divider */}
          <div className="h-px w-12 bg-gray-300" />

          {/* Step 2 */}
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold",
                currentStep >= 2
                  ? "bg-blue-600 text-white"
                  : "bg-gray-200 text-gray-600"
              )}
            >
              2
            </div>
            <span
              className={cn(
                "text-sm font-medium",
                currentStep >= 2 ? "text-blue-600" : "text-gray-500"
              )}
            >
              Pratinjau
            </span>
          </div>

          {/* Divider */}
          <div className="h-px w-12 bg-gray-300" />

          {/* Step 3 */}
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold",
                currentStep >= 3
                  ? "bg-blue-600 text-white"
                  : "bg-gray-200 text-gray-600"
              )}
            >
              3
            </div>
            <span
              className={cn(
                "text-sm font-medium",
                currentStep >= 3 ? "text-blue-600" : "text-gray-500"
              )}
            >
              Selesai
            </span>
          </div>
        </div>

        {/* Content based on step */}
        <div className="py-4">
          {currentStep === 1 && (
            <div className="space-y-4">
              {/* Upload Area */}
              <div
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                className={cn(
                  "flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors",
                  isDragging
                    ? "border-blue-500 bg-blue-50"
                    : "border-gray-300 bg-gray-50"
                )}
              >
                <Upload className="h-12 w-12 text-blue-500 mb-4" />
                <p className="text-center text-sm text-gray-900 font-medium mb-1">
                  Tarik & Lepas file Excel di sini atau{" "}
                  <label htmlFor="file-upload" className="text-blue-600 cursor-pointer hover:text-blue-700">
                    Klik untuk memilih
                  </label>
                </p>
                <p className="text-xs text-gray-500">
                  Format yang didukung: .xlsx, .xls (Maks. 5MB)
                </p>
                <input
                  id="file-upload"
                  type="file"
                  accept=".xlsx,.xls,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                  onChange={handleFileInputChange}
                  className="hidden"
                />
              </div>

              {/* Selected File */}
              {selectedFile && (
                <div className="flex items-center gap-3 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <FileSpreadsheet className="h-8 w-8 text-blue-600" />
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-900">
                      {selectedFile.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {(selectedFile.size / 1024).toFixed(2)} KB
                    </p>
                  </div>
                  <button
                    onClick={() => setSelectedFile(null)}
                    className="text-red-600 hover:text-red-700 text-sm font-medium"
                  >
                    Hapus
                  </button>
                </div>
              )}

              {/* Download Template */}
              <div className="flex items-center gap-3 p-4 bg-gray-50 border border-gray-200 rounded-lg">
                <FileSpreadsheet className="h-6 w-6 text-green-600" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-900">
                    Belum punya formatnya?
                  </p>
                  <p className="text-xs text-gray-500">
                    Gunakan template resmi kami
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDownloadTemplate}
                  className="flex items-center gap-2"
                >
                  <Download className="h-4 w-4" />
                  Unduh Template Excel
                </Button>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <div className="space-y-4">
              <div className="p-8 border-2 border-dashed border-gray-300 rounded-lg text-center">
                <FileCheck className="h-12 w-12 text-blue-500 mx-auto mb-4" />
                <p className="text-sm text-gray-900 font-medium mb-2">
                  Pratinjau Data
                </p>
                <p className="text-xs text-gray-500">
                  Menampilkan pratinjau dari file: {selectedFile?.name}
                </p>
                {/* TODO: Add data preview table here */}
                <div className="mt-4 p-4 bg-gray-50 rounded text-left">
                  <p className="text-xs text-gray-600">
                    Total pegawai: <strong>25</strong>
                  </p>
                  <p className="text-xs text-gray-600">
                    Valid: <strong className="text-green-600">23</strong>
                  </p>
                  <p className="text-xs text-gray-600">
                    Error: <strong className="text-red-600">2</strong>
                  </p>
                </div>
              </div>
            </div>
          )}

          {currentStep === 3 && (
            <div className="space-y-4">
              <div className="p-8 border-2 border-green-200 bg-green-50 rounded-lg text-center">
                <div className="h-12 w-12 bg-green-500 rounded-full flex items-center justify-center mx-auto mb-4">
                  <FileCheck className="h-6 w-6 text-white" />
                </div>
                <p className="text-sm text-gray-900 font-semibold mb-2">
                  Import Berhasil!
                </p>
                <p className="text-xs text-gray-600">
                  23 pegawai berhasil diimport ke sistem
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="flex justify-end gap-3 pt-4 border-t">
          <Button variant="outline" onClick={handleCancel}>
            Batal
          </Button>
          {currentStep < 3 ? (
            <Button
              onClick={handleNext}
              disabled={currentStep === 1 && !selectedFile}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              Lanjutkan
            </Button>
          ) : (
            <Button
              onClick={handleComplete}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              Selesai
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}