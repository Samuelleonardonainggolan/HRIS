"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AlertTriangle, CheckCircle, Ban } from "lucide-react";
import { cn } from "@/lib/utils";

interface ConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  variant?: "default" | "destructive";
  icon?: "warning" | "activate" | "deactivate";
}

export function ConfirmationDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmText = "Ya, Lanjutkan",
  cancelText = "Batal",
  onConfirm,
  variant = "default",
  icon = "warning",
}: ConfirmationDialogProps) {
  const handleConfirm = () => {
    onConfirm();
    onOpenChange(false);
  };

  const getIcon = () => {
    switch (icon) {
      case "activate":
        return <CheckCircle className="h-6 w-6 text-emerald-600" />;
      case "deactivate":
        return <Ban className="h-6 w-6 text-amber-600" />;
      default:
        return <AlertTriangle className="h-6 w-6 text-amber-600" />;
    }
  };

  const getIconBackground = () => {
    switch (icon) {
      case "activate":
        return "bg-emerald-100";
      case "deactivate":
        return "bg-amber-100";
      default:
        return "bg-amber-100";
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <div className="flex items-center gap-4">
            <div className={cn("flex h-10 w-10 items-center justify-center rounded-full", getIconBackground())}>
              {getIcon()}
            </div>
            <div>
              <DialogTitle className="text-left">{title}</DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-left pt-2">
            {description}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="mt-2 sm:mt-0"
          >
            {cancelText}
          </Button>
          <Button
            variant={variant === "destructive" ? "destructive" : "default"}
            onClick={handleConfirm}
            className={cn(
              variant === "default" && "bg-blue-600 hover:bg-blue-700 text-white"
            )}
          >
            {confirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
