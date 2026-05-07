"use client";

import * as DialogPrimitive from "@radix-ui/react-dialog";
import {
  X, Phone, MapPin, Briefcase, Building2,
  Calendar, GraduationCap, BookOpen, UserCheck,
} from "lucide-react";
import { User as UserType } from "@/lib/api/auth";

interface ProfileModalProps {
  open: boolean;
  onClose: () => void;
  user: UserType | null;
}

/* ─── Single info field ──────────────────────────────────────────────────── */
function Field({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType;
  label: string;
  value?: string | null;
}) {
  if (!value) return null;
  return (
    <div className="flex items-start gap-3 rounded-xl bg-gray-50 px-4 py-3.5 transition-colors hover:bg-gray-100/80">
      <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white shadow-sm ring-1 ring-gray-200">
        <Icon className="h-4 w-4 text-gray-500" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-xs font-semibold uppercase tracking-wide text-gray-400">
          {label}
        </p>
        <p className="mt-0.5 text-sm font-semibold text-gray-800 break-words leading-snug">
          {value}
        </p>
      </div>
    </div>
  );
}

/* ─── Section heading ────────────────────────────────────────────────────── */
function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3 mb-3">
      <p className="text-xs font-bold uppercase tracking-widest text-gray-400 whitespace-nowrap">
        {children}
      </p>
      <div className="flex-1 h-px bg-gray-200" />
    </div>
  );
}

/* ─── Main component ─────────────────────────────────────────────────────── */
export function ProfileModal({ open, onClose, user }: ProfileModalProps) {
  if (!user) return null;

  const getInitials = () => {
    if (!user.full_name) return "U";
    const parts = user.full_name.trim().split(" ");
    return parts.length >= 2
      ? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
      : parts[0][0].toUpperCase();
  };

  const formatDate = (dateStr?: string | null) => {
    if (!dateStr) return null;
    try {
      return new Date(dateStr).toLocaleDateString("id-ID", {
        day: "numeric",
        month: "long",
        year: "numeric",
      });
    } catch {
      return dateStr;
    }
  };

  const roleLabel: Record<string, string> = {
    manager_hr:         "Manager HR",
    manager_departemen: "Manager Departemen",
    admin_departemen:   "Admin Departemen",
    staf:               "Staf",
  };

  const dept = user.department || user.department_name;
  const pos  = user.position   || user.position_name;
  const nik  = user.nik        || user.payroll_number;

  return (
    <DialogPrimitive.Root open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogPrimitive.Portal>
        {/* Overlay */}
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />

        {/* Panel */}
        <DialogPrimitive.Content
          className="
            fixed left-1/2 top-1/2 z-50
            w-[calc(100vw-2rem)] max-w-[600px]
            -translate-x-1/2 -translate-y-1/2
            overflow-hidden rounded-2xl bg-white shadow-2xl ring-1 ring-black/5
            duration-200
            data-[state=open]:animate-in  data-[state=closed]:animate-out
            data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0
            data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95
            data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]
            data-[state=open]:slide-in-from-left-1/2  data-[state=open]:slide-in-from-top-[48%]
          "
        >
          {/* ── Hero ── */}
          <div className="relative bg-gradient-to-br from-orange-500 via-orange-400 to-amber-400 px-8 pb-12 pt-10">
            {/* Close button */}
            <DialogPrimitive.Close className="absolute right-5 top-5 flex h-8 w-8 items-center justify-center rounded-full bg-white/20 text-white transition-colors hover:bg-white/35 focus:outline-none focus:ring-2 focus:ring-white/50">
              <X className="h-4 w-4" />
              <span className="sr-only">Tutup</span>
            </DialogPrimitive.Close>

            <div className="flex flex-col items-center gap-4">
              {/* Avatar */}
              <div className="relative">
                <div className="h-24 w-24 rounded-full bg-white/20 p-1 shadow-2xl ring-2 ring-white/50">
                  {user.avatar ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={user.avatar}
                      alt={user.full_name}
                      className="h-full w-full rounded-full object-cover"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center rounded-full bg-white/30 text-3xl font-bold text-white">
                      {getInitials()}
                    </div>
                  )}
                </div>
                {user.is_active && (
                  <span className="absolute bottom-1 right-1 h-5 w-5 rounded-full border-2 border-white bg-emerald-400 shadow-md" />
                )}
              </div>

              {/* Name & email */}
              <div className="text-center">
                <DialogPrimitive.Title className="text-2xl font-bold text-white drop-shadow-sm">
                  {user.full_name}
                </DialogPrimitive.Title>
                <p className="mt-1 text-sm text-white/80">{user.email}</p>
              </div>

              {/* Badges */}
              <div className="flex flex-wrap justify-center gap-2">
                <span className="inline-flex items-center rounded-full bg-white/20 px-4 py-1 text-sm font-semibold text-white ring-1 ring-white/30 backdrop-blur-sm">
                  {roleLabel[user.role] ?? user.role}
                </span>
                {user.is_active && (
                  <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/30 px-4 py-1 text-sm font-semibold text-white ring-1 ring-emerald-300/40 backdrop-blur-sm">
                    <span className="h-2 w-2 rounded-full bg-emerald-300" />
                    Aktif
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* ── Quick stats strip ── */}
          {(nik || dept || pos) && (
            <div className="grid grid-cols-3 divide-x divide-gray-100 border-b border-gray-100 bg-gray-50/70">
              {nik && (
                <div className="flex flex-col items-center py-4 px-3">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-gray-400">NIK</p>
                  <p className="mt-1 text-sm font-bold text-gray-700 truncate max-w-full">{nik}</p>
                </div>
              )}
              {dept && (
                <div className="flex flex-col items-center py-4 px-3">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-gray-400">Departemen</p>
                  <p className="mt-1 text-sm font-bold text-gray-700 text-center leading-tight line-clamp-2">{dept}</p>
                </div>
              )}
              {pos && (
                <div className="flex flex-col items-center py-4 px-3">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-gray-400">Jabatan</p>
                  <p className="mt-1 text-sm font-bold text-gray-700 text-center leading-tight line-clamp-2">{pos}</p>
                </div>
              )}
            </div>
          )}

          {/* ── Detail sections ── */}
          <div className="max-h-[420px] overflow-y-auto px-7 py-6 space-y-6">

            {/* Informasi Pekerjaan */}
            {(user.employment_status || user.join_date || dept || pos) && (
              <section>
                <SectionTitle>Informasi Pekerjaan</SectionTitle>
                <div className="grid grid-cols-2 gap-2.5">
                  <Field icon={UserCheck} label="Status Kepegawaian" value={user.employment_status} />
                  <Field icon={Calendar}  label="Tanggal Bergabung"  value={formatDate(user.join_date)} />
                  <Field icon={Building2} label="Departemen"         value={dept} />
                  <Field icon={Briefcase} label="Jabatan"            value={pos} />
                </div>
              </section>
            )}

            {/* Informasi Pribadi */}
            {(user.phone || user.birth_date || user.religion || user.last_education || user.address) && (
              <section>
                <SectionTitle>Informasi Pribadi</SectionTitle>
                <div className="grid grid-cols-2 gap-2.5">
                  <Field icon={Phone}         label="Telepon"             value={user.phone} />
                  <Field icon={Calendar}      label="Tanggal Lahir"       value={formatDate(user.birth_date)} />
                  <Field icon={BookOpen}      label="Agama"               value={user.religion} />
                  <Field icon={GraduationCap} label="Pendidikan Terakhir" value={user.last_education} />
                </div>
                {user.address && (
                  <div className="mt-2.5">
                    <Field icon={MapPin} label="Alamat" value={user.address} />
                  </div>
                )}
              </section>
            )}
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
