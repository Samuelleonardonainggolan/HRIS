"use client";

import { useState, useEffect } from "react";
import { Bell, LogOut, User, ChevronDown } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuth } from "@/contexts/AuthContext";
import { ProfileModal } from "@/components/profile-modal";
import { useRouter } from "next/navigation";
import { notificationsApi, NotificationResponse } from "@/lib/api/notifications";
import { authService } from "@/lib/api/auth";
import toast from "react-hot-toast";

export function Header() {
  const [currentTime, setCurrentTime] = useState<Date>(new Date());
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [notifications, setNotifications] = useState<NotificationResponse[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const { user, logout } = useAuth();
  const router = useRouter();

  // Update waktu setiap detik
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const handleLogout = async () => {
    await logout();
  };

  const fetchNotifications = async () => {
    try {
      const data = await notificationsApi.getNotifications(10);
      setNotifications(data);
      const count = await notificationsApi.getUnreadCount();
      setUnreadCount(count);
    } catch (err) {
      console.error("Gagal memuat notifikasi:", err);
    }
  };

  const handleMarkAsRead = async (n: NotificationResponse) => {
    try {
      if (!n.is_read) {
        await notificationsApi.markAsRead(n.id);
        fetchNotifications();
      }
      setShowNotifications(false);
      
      if (n.reference_id) {
        if (user?.role === "manager_hr") {
          router.push(`/dashboard/manager-hr/persetujuan-izin-cuti`);
        } else if (user?.role === "manager_departemen") {
          router.push(`/dashboard/manager-dept/persetujuan-izin-cuti`);
        } else {
          router.push(`/dashboard/staff`);
        }
      }
    } catch (err) {
      console.error("Gagal menandai notifikasi dibaca:", err);
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await notificationsApi.markAllAsRead();
      fetchNotifications();
      toast.success("Semua notifikasi ditandai dibaca");
    } catch (err) {
      console.error("Gagal menandai semua dibaca:", err);
    }
  };

  const formatTimeAgo = (dateStr: string) => {
    try {
      const date = new Date(dateStr);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMins / 60);
      const diffDays = Math.floor(diffHours / 24);

      if (diffMins < 1) return "Baru saja";
      if (diffMins < 60) return `${diffMins} mnt lalu`;
      if (diffHours < 24) return `${diffHours} jam lalu`;
      return `${diffDays} hari lalu`;
    } catch {
      return "";
    }
  };

  useEffect(() => {
    if (!user) return;

    fetchNotifications();

    let eventSource: EventSource | null = null;
    
    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const base = apiBase.startsWith("http")
        ? apiBase
        : `${window.location.origin}${apiBase === "/" ? "" : apiBase}`;
      const token = authService.getAccessToken();
      
      if (token) {
        eventSource = new EventSource(`${base}/api/v1/realtime/connect?token=${token}`);
        
        eventSource.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === "notification_created") {
              toast.success(
                <div className="flex flex-col gap-0.5">
                  <span className="font-semibold text-gray-900">{data.payload?.title || "Notifikasi Baru"}</span>
                  <span className="text-xs text-gray-600">{data.payload?.message || ""}</span>
                </div>,
                { duration: 5000, icon: "🔔" }
              );
              fetchNotifications();
            }
          } catch (e) {
            console.error("Failed to parse SSE payload:", e);
          }
        };

        eventSource.onerror = (err) => {
          console.error("SSE connection error, attempting reconnect...", err);
        };
      }
    } catch (err) {
      console.error("Error setting up EventSource", err);
    }

    return () => {
      if (eventSource) {
        eventSource.close();
      }
    };
  }, [user]);

  // Format tanggal: "Senin, 23 Okt"
  const currentDate = currentTime.toLocaleDateString("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "short",
  });

  // Format waktu: "09:45:30 WIB"
  const formattedTime =
    currentTime.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }) + " WIB";

  // Get user initials
  const getUserInitials = () => {
    if (!user?.full_name) return "U";
    const names = user.full_name.split(" ");
    if (names.length >= 2) {
      return names[0][0] + names[1][0];
    }
    return names[0][0];
  };

  return (
    <>
      <header className="border-b border-gray-200 bg-white px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Search Bar */}
          <div className="flex-1 max-w-xl">
            {/* Search can be added here if needed */}
          </div>

          {/* Right Section */}
          <div className="flex items-center gap-4">
            {/* Date & Time */}
            <div className="text-right">
              <p className="text-sm font-medium text-gray-900">{currentDate}</p>
              <p className="text-xs text-gray-500">Waktu Lokal: {formattedTime}</p>
            </div>

            {/* Notification Bell */}
            <div className="relative">
              <button 
                onClick={() => setShowNotifications(!showNotifications)}
                className="relative rounded-lg p-2 hover:bg-gray-100 focus:outline-none transition-colors"
                aria-label="Notifikasi"
              >
                <Bell className="h-5 w-5 text-gray-600" />
                {unreadCount > 0 && (
                  <span className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white ring-2 ring-white animate-pulse">
                    {unreadCount > 9 ? "9+" : unreadCount}
                  </span>
                )}
              </button>

              {/* Notification Popover Dropdown */}
              {showNotifications && (
                <>
                  {/* Backdrop */}
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowNotifications(false)}
                  />

                  {/* Popover Card */}
                  <div className="absolute right-0 mt-2 w-80 sm:w-96 bg-white rounded-xl shadow-2xl border border-gray-150 py-2 z-20 transition-all duration-200 ease-out origin-top-right">
                    <div className="flex items-center justify-between px-4 py-2 border-b border-gray-100">
                      <span className="text-sm font-semibold text-gray-900 flex items-center gap-1.5">
                        <Bell className="h-4 w-4 text-emerald-600" />
                        Notifikasi
                      </span>
                      {unreadCount > 0 && (
                        <button
                          onClick={handleMarkAllAsRead}
                          className="text-xs text-emerald-600 hover:text-emerald-700 font-medium transition-colors hover:underline cursor-pointer"
                        >
                          Tandai semua dibaca
                        </button>
                      )}
                    </div>

                    {/* Notifications List */}
                    <div className="max-h-[360px] overflow-y-auto divide-y divide-gray-50">
                      {notifications.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
                          <div className="h-10 w-10 rounded-full bg-gray-50 flex items-center justify-center mb-2">
                            <Bell className="h-5 w-5 text-gray-400" />
                          </div>
                          <p className="text-sm font-medium text-gray-500">Tidak ada notifikasi baru</p>
                          <p className="text-xs text-gray-400 mt-0.5">Semua info terbaru Anda akan muncul di sini</p>
                        </div>
                      ) : (
                        notifications.map((n) => (
                          <div
                            key={n.id}
                            onClick={() => handleMarkAsRead(n)}
                            className={`flex flex-col gap-1 px-4 py-3 cursor-pointer transition-colors relative hover:bg-gray-50 ${
                              !n.is_read ? "bg-emerald-50/40 hover:bg-emerald-50/70" : ""
                            }`}
                          >
                            {!n.is_read && (
                              <span className="absolute left-2 top-4 h-1.5 w-1.5 rounded-full bg-emerald-500" />
                            )}
                            <div className="flex items-start justify-between gap-2">
                              <span className={`text-xs font-semibold ${!n.is_read ? "text-gray-900" : "text-gray-700"}`}>
                                {n.title}
                              </span>
                              <span className="text-[10px] text-gray-400 whitespace-nowrap">
                                {formatTimeAgo(n.created_at)}
                              </span>
                            </div>
                            <p className="text-xs text-gray-600 leading-relaxed break-words">
                              {n.message}
                            </p>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                </>
              )}
            </div>

            {/* User Menu */}
            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-2 rounded-lg p-2 hover:bg-gray-100"
              >
                <Avatar>
                  <AvatarFallback className="bg-orange-400 text-white">
                    {getUserInitials()}
                  </AvatarFallback>
                </Avatar>
                <ChevronDown className="h-4 w-4 text-gray-600" />
              </button>

              {/* Dropdown Menu */}
              {showUserMenu && (
                <>
                  {/* Backdrop */}
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowUserMenu(false)}
                  />

                  {/* Menu */}
                  <div className="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-20">
                    <div className="px-4 py-3 border-b border-gray-200">
                      <p className="text-sm font-medium text-gray-900">
                        {user?.full_name}
                      </p>
                      <p className="text-sm text-gray-500">{user?.email}</p>
                      <p className="text-xs text-gray-400 mt-1">
                        {user?.department} - {user?.position}
                      </p>
                    </div>

                    <button
                      onClick={() => {
                        setShowUserMenu(false);
                        setShowProfile(true);
                      }}
                      className="w-full flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      <User className="mr-3 h-4 w-4" />
                      Profile Saya
                    </button>

                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                    >
                      <LogOut className="mr-3 h-4 w-4" />
                      Logout
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Profile Modal */}
      <ProfileModal
        open={showProfile}
        onClose={() => setShowProfile(false)}
        user={user}
      />
    </>
  );
}
