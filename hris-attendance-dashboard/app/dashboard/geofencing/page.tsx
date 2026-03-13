'use client';

import { useState, useEffect } from 'react';
import { geofenceApi } from '@/lib/api/geofence';
import { Geofence, CreateGeofenceRequest } from '@/types/geofence';
import GeofencingMap from '@/components/geofencing/GeofencingMap';
import LocationList from '@/components/geofencing/LocationList';
import LocationConfiguration from '@/components/geofencing/LocationConfiguration';
import { MapPinIcon, PlusIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';

export default function GeofencingPage() {
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [selectedGeofence, setSelectedGeofence] = useState<Geofence | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // UI State
  const [showConfiguration, setShowConfiguration] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [newLocation, setNewLocation] = useState<{ lat: number; lng: number } | null>(null);

  // Fetch geofences on mount
  useEffect(() => {
    loadGeofences();
  }, []);

  const loadGeofences = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await geofenceApi.getAll();
      setGeofences(data);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load geofences';
      console.error('Error loading geofences:', err);
      setError(message);
      toast.error('Failed to load geofences');
    } finally {
      setIsLoading(false);
    }
  };

  const handleMapClick = (lat: number, lng: number) => {
    setNewLocation({ lat, lng });
    setSelectedGeofence(null);
    setIsEditing(false);
    setShowConfiguration(true);
  };

  const handleGeofenceSelect = (geofence: Geofence) => {
    setSelectedGeofence(geofence);
    setNewLocation(null);
    setShowConfiguration(false);
  };

  const handleEdit = (geofence: Geofence) => {
    setSelectedGeofence(geofence);
    setIsEditing(true);
    setNewLocation(null);
    setShowConfiguration(true);
  };

  const handleDelete = async (geofence: Geofence) => {
    if (!confirm(`Are you sure you want to delete "${geofence.name}"?`)) {
      return;
    }

    try {
      await geofenceApi.delete(geofence.id);
      toast.success('Geofence deleted successfully');
      await loadGeofences();
      if (selectedGeofence?.id === geofence.id) {
        setSelectedGeofence(null);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to delete geofence';
      console.error('Error deleting geofence:', err);
      toast.error(message);
    }
  };

  const handleSave = async (data: CreateGeofenceRequest) => {
    try {
      if (isEditing && selectedGeofence) {
        // Update existing
        await geofenceApi.update(selectedGeofence.id, data);
        toast.success('Geofence updated successfully');
      } else {
        // Create new
        await geofenceApi.create(data);
        toast.success('Geofence created successfully');
      }
      
      await loadGeofences();
      setShowConfiguration(false);
      setIsEditing(false);
      setSelectedGeofence(null);
      setNewLocation(null);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to save geofence';
      console.error('Error saving geofence:', err);
      toast.error(message);
    }
  };

  const handleCancel = () => {
    setShowConfiguration(false);
    setIsEditing(false);
    setSelectedGeofence(null);
    setNewLocation(null);
  };

  const handleAddNew = () => {
    setSelectedGeofence(null);
    setNewLocation(null);
    setIsEditing(false);
    setShowConfiguration(true);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-3">
                <MapPinIcon className="w-8 h-8 text-blue-600" />
                Manajemen Geofencing
              </h1>
              <p className="text-sm text-gray-500 mt-1">
                Kelola lokasi dan radius geofencing untuk absensi karyawan
              </p>
            </div>
            <button
              onClick={handleAddNew}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <PlusIcon className="w-5 h-5" />
              Tambah Lokasi
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Map Section */}
          <div className="lg:col-span-8">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">🗺️ Peta Lokasi</h2>
                <div className="text-sm text-gray-500">
                  Klik pada peta untuk menambah lokasi baru
                </div>
              </div>
              
              {isLoading ? (
                <div className="h-[600px] flex items-center justify-center bg-gray-100 rounded-lg">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-600">Loading geofences...</p>
                  </div>
                </div>
              ) : error ? (
                <div className="h-[600px] flex items-center justify-center bg-red-50 rounded-lg">
                  <div className="text-center p-6">
                    <p className="text-red-600 mb-4">{error}</p>
                    <button
                      onClick={loadGeofences}
                      className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
                    >
                      Retry
                    </button>
                  </div>
                </div>
              ) : (
                <div className="h-[600px]">
                  <GeofencingMap
                    geofences={geofences}
                    selectedGeofence={selectedGeofence}
                    onGeofenceSelect={handleGeofenceSelect}
                    onMapClick={handleMapClick}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-4 space-y-6">
            {/* Configuration Panel (when active) */}
            {showConfiguration && (
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
                <div className="bg-blue-50 px-4 py-3 border-b border-blue-100 flex items-center justify-between">
                  <h3 className="font-semibold text-gray-900">
                    {isEditing ? '✏️ Edit Lokasi' : '➕ Tambah Lokasi Baru'}
                  </h3>
                  <button
                    onClick={handleCancel}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <XMarkIcon className="w-5 h-5" />
                  </button>
                </div>
                <div className="p-4">
                  <LocationConfiguration
                    selectedGeofence={isEditing ? selectedGeofence : undefined}
                    latitude={newLocation?.lat}
                    longitude={newLocation?.lng}
                    onSave={handleSave}
                    onCancel={handleCancel}
                  />
                </div>
              </div>
            )}

            {/* Location List */}
            <LocationList
              geofences={geofences}
              selectedGeofence={selectedGeofence}
              onSelect={handleGeofenceSelect}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />

            {/* Info Panel */}
            <div className="bg-blue-50 rounded-lg border border-blue-100 p-4">
              <h4 className="font-semibold text-blue-900 mb-2">💡 Cara Penggunaan</h4>
              <ul className="text-sm text-blue-800 space-y-2">
                <li className="flex items-start gap-2">
                  <span className="font-bold">1.</span>
                  <span>Klik pada peta untuk menambah lokasi baru</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="font-bold">2.</span>
                  <span>Atur nama, radius, dan konfigurasi lokasi</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="font-bold">3.</span>
                  <span>Klik lokasi di daftar untuk melihat detail</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="font-bold">4.</span>
                  <span>Gunakan tombol edit/hapus untuk mengelola</span>
                </li>
              </ul>
            </div>

            {/* Statistics */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
              <h4 className="font-semibold text-gray-900 mb-3">📊 Statistik</h4>
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Total Lokasi:</span>
                  <span className="font-semibold text-gray-900">{geofences.length}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Aktif:</span>
                  <span className="font-semibold text-green-600">
                    {geofences.filter(g => g.is_active).length}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Tidak Aktif:</span>
                  <span className="font-semibold text-red-600">
                    {geofences.filter(g => !g.is_active).length}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}