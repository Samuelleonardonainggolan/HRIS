"use client";

import { useState, useEffect } from "react";
import { Search, Bell, LogOut, User, ChevronDown } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useAuth } from "@/contexts/AuthContext";

export function Header() {
  const [currentTime, setCurrentTime] = useState<Date>(new Date());
  const [showUserMenu, setShowUserMenu] = useState(false);
  const { user, logout } = useAuth();

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
          <button className="relative rounded-lg p-2 hover:bg-gray-100">
            <Bell className="h-5 w-5 text-gray-600" />
            <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500"></span>
          </button>

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
                ></div>

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

                  <button className="w-full flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
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
  );
}