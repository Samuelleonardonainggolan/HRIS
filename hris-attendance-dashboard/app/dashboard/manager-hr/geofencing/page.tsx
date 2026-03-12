'use client';

import { useState, useEffect } from 'react';
import { geofenceApi } from '@/lib/api/geofence';
import { Geofence } from '@/types/geofence';
import GeofencingMap from '@/components/geofencing/GeofencingMap';
import { MapPinIcon, Cog6ToothIcon, ArrowPathIcon } from '@heroicons/react/24/outline';
import { toast, Toaster } from 'react-hot-toast';

export default function GeofencingPage() {
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [selectedGeofence, setSelectedGeofence] = useState<Geofence | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Configuration state
  const [showConfig, setShowConfig] = useState(false);
  const [configData, setConfigData] = useState({
    id: '',
    name: '',
    latitude: -6.2088,
    longitude: 106.8456,
    radius: 150,
    color: '#3B82F6',
    icon: '📍',
  });

  // ✅ ADD: Preview geofence untuk live preview saat edit
  const [previewGeofence, setPreviewGeofence] = useState<Geofence | null>(null);

  useEffect(() => {
    loadGeofences();
  }, []);

  // ✅ ADD: Update preview saat configData berubah
  useEffect(() => {
    if (showConfig) {
      setPreviewGeofence({
        id: configData.id || 'preview',
        name: configData.name || 'Preview Location',
        description: '',
        latitude: configData.latitude,
        longitude: configData.longitude,
        address: '',
        radius: configData.radius,
        radius_unit: 'meter',
        icon: configData.icon,
        color: configData.color,
        applies_to: 'all',
        is_active: true,
        created_by: '',
        created_by_name: '',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      });
    } else {
      setPreviewGeofence(null);
    }
  }, [configData, showConfig]);

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
      toast.error('Gagal memuat data geofence');
    } finally {
      setIsLoading(false);
    }
  };

  const handleMapClick = (lat: number, lng: number) => {
    setConfigData({
      id: '',
      name: '',
      latitude: lat,
      longitude: lng,
      radius: 150,
      color: '#3B82F6',
      icon: '📍',
    });
    setSelectedGeofence(null);
    setShowConfig(true);
  };

  const handleGeofenceSelect = (geofence: Geofence) => {
    setSelectedGeofence(geofence);
    setConfigData({
      id: geofence.id,
      name: geofence.name,
      latitude: geofence.latitude,
      longitude: geofence.longitude,
      radius: geofence.radius,
      color: geofence.color,
      icon: geofence.icon,
    });
    setShowConfig(true);
  };

  const handleSaveChanges = async () => {
    try {
      if (configData.id) {
        // Update existing
        await geofenceApi.update(configData.id, {
          name: configData.name,
          description: '',
          latitude: configData.latitude,
          longitude: configData.longitude,
          address: '',
          radius: configData.radius,
          icon: configData.icon,
          color: configData.color,
          applies_to: 'all',
        });
        toast.success('Lokasi berhasil diperbarui');
      } else {
        // Create new
        await geofenceApi.create({
          name: configData.name,
          description: '',
          latitude: configData.latitude,
          longitude: configData.longitude,
          address: '',
          radius: configData.radius,
          icon: configData.icon,
          color: configData.color,
          applies_to: 'all',
        });
        toast.success('Lokasi berhasil ditambahkan');
      }
      
      setShowConfig(false);
      setPreviewGeofence(null);
      await loadGeofences();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Gagal menyimpan perubahan';
      console.error('Error saving geofence:', err);
      toast.error(message);
    }
  };

  const handleResetToDefault = () => {
    setConfigData({
      id: '',
      name: '',
      latitude: -6.2088,
      longitude: 106.8456,
      radius: 150,
      color: '#3B82F6',
      icon: '📍',
    });
  };

  const handleCancelConfig = () => {
    setShowConfig(false);
    setPreviewGeofence(null);
    setSelectedGeofence(null);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Toaster position="top-right" />

      {/* Header */}
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center gap-3">
          <MapPinIcon className="w-6 h-6 text-gray-700" />
          <h1 className="text-xl font-semibold text-gray-900">Manajemen Biometrik</h1>
        </div>
      </div>

      {/* Main Content */}
      <div className="p-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left: Map */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
              {/* Map Header */}
              <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between bg-gray-50">
                <div className="flex items-center gap-2">
                  <MapPinIcon className="w-5 h-5 text-blue-600" />
                  <span className="font-medium text-gray-700">Pratinjau Peta</span>
                </div>
                <div className="flex items-center gap-2">
                  <button 
                    className="p-1.5 hover:bg-gray-200 rounded"
                    title="Settings"
                  >
                    <Cog6ToothIcon className="w-5 h-5 text-gray-600" />
                  </button>
                  <button 
                    className="p-1.5 hover:bg-gray-200 rounded"
                    onClick={loadGeofences}
                    title="Refresh"
                  >
                    <ArrowPathIcon className="w-5 h-5 text-gray-600" />
                  </button>
                </div>
              </div>

              {/* Map */}
              <div className="h-[500px]">
                {isLoading ? (
                  <div className="h-full flex items-center justify-center bg-gray-100">
                    <div className="text-center">
                      <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600 mx-auto mb-3"></div>
                      <p className="text-sm text-gray-600">Memuat peta...</p>
                    </div>
                  </div>
                ) : error ? (
                  <div className="h-full flex items-center justify-center bg-red-50">
                    <div className="text-center p-4">
                      <p className="text-red-600 mb-3">{error}</p>
                      <button
                        onClick={loadGeofences}
                        className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm"
                      >
                        Coba Lagi
                      </button>
                    </div>
                  </div>
                ) : (
                  <GeofencingMap
                    geofences={[
                      ...geofences,
                      ...(previewGeofence ? [previewGeofence] : []) // ✅ ADD: Show preview
                    ]}
                    selectedGeofence={previewGeofence || selectedGeofence}
                    onGeofenceSelect={handleGeofenceSelect}
                    onMapClick={handleMapClick}
                  />
                )}
              </div>
            </div>
          </div>

          {/* Right: Configuration Panel */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200">
              {/* Config Header */}
              <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Cog6ToothIcon className="w-5 h-5 text-blue-600" />
                    <span className="font-medium text-gray-700">Konfigurasi Lokasi</span>
                  </div>
                  {showConfig && (
                    <button
                      onClick={handleCancelConfig}
                      className="text-gray-500 hover:text-gray-700"
                    >
                      ✕
                    </button>
                  )}
                </div>
              </div>

              {/* Config Form */}
              <div className="p-4 space-y-4">
                {/* Nama Lokasi */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">
                    Nama Lokasi
                  </label>
                  <input
                    type="text"
                    value={configData.name}
                    onChange={(e) => setConfigData({ ...configData, name: e.target.value })}
                    placeholder="Hotel Labosta"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                {/* Latitude & Longitude */}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                      Latitude
                    </label>
                    <input
                      type="number"
                      step="0.000001"
                      value={configData.latitude}
                      onChange={(e) => setConfigData({ ...configData, latitude: parseFloat(e.target.value) || 0 })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                      Longitude
                    </label>
                    <input
                      type="number"
                      step="0.000001"
                      value={configData.longitude}
                      onChange={(e) => setConfigData({ ...configData, longitude: parseFloat(e.target.value) || 0 })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>

                {/* Radius */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">
                    Radius (Meter)
                  </label>
                  <div className="flex items-center gap-3">
                    <input
                      type="range"
                      min="10"
                      max="1000"
                      step="10"
                      value={configData.radius}
                      onChange={(e) => setConfigData({ ...configData, radius: parseInt(e.target.value) })}
                      className="flex-1"
                    />
                    <div className="w-20 px-3 py-2 bg-blue-50 border border-blue-200 rounded-lg text-center">
                      <span className="text-sm font-semibold text-blue-700">{configData.radius}m</span>
                    </div>
                  </div>
                  <input
                    type="number"
                    value={configData.radius}
                    onChange={(e) => setConfigData({ ...configData, radius: parseInt(e.target.value) || 10 })}
                    min="10"
                    max="10000"
                    className="w-full mt-2 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                {/* ✅ ADD: Preview indicator */}
                {showConfig && (
                  <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
                    <p className="text-xs text-blue-800">
                      👁️ Preview aktif - Lihat perubahan di peta secara real-time
                    </p>
                  </div>
                )}

                {/* Action Buttons */}
                <div className="pt-4 space-y-2">
                  <button
                    onClick={handleSaveChanges}
                    disabled={!configData.name}
                    className="w-full py-2.5 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors text-sm disabled:bg-gray-400 disabled:cursor-not-allowed"
                  >
                    {configData.id ? 'Update Lokasi' : 'Simpan Lokasi'}
                  </button>
                  <button
                    onClick={handleResetToDefault}
                    className="w-full py-2.5 border border-gray-300 text-gray-700 rounded-lg font-medium hover:bg-gray-50 transition-colors text-sm"
                  >
                    Reset ke Default
                  </button>
                  {showConfig && (
                    <button
                      onClick={handleCancelConfig}
                      className="w-full py-2.5 border border-gray-300 text-gray-700 rounded-lg font-medium hover:bg-gray-50 transition-colors text-sm"
                    >
                      Batal
                    </button>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Lokasi Terdaftar Table */}
        <div className="mt-6 bg-white rounded-lg shadow-sm border border-gray-200">
          {/* Table Header */}
          <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
            <h3 className="font-semibold text-gray-900">Lokasi Terdaftar</h3>
            <p className="text-sm text-gray-600 mt-0.5">Daftar semua koordinat geofencing yang aktif</p>
          </div>

          {/* Table */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Lokasi
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Alamat Lengkap
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Koordinat
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Radius
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                    Aksi
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {isLoading ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                      <p className="mt-2 text-sm">Memuat data...</p>
                    </td>
                  </tr>
                ) : geofences.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                      <MapPinIcon className="w-12 h-12 mx-auto text-gray-400 mb-2" />
                      <p className="text-sm">Belum ada lokasi terdaftar</p>
                      <p className="text-xs text-gray-400 mt-1">Klik pada peta untuk menambah lokasi baru</p>
                    </td>
                  </tr>
                ) : (
                  geofences.map((geofence) => (
                    <tr 
                      key={geofence.id} 
                      className={`hover:bg-gray-50 transition-colors ${
                        selectedGeofence?.id === geofence.id ? 'bg-blue-50' : ''
                      }`}
                    >
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <div 
                            className="w-8 h-8 rounded flex items-center justify-center text-lg" 
                            style={{ backgroundColor: `${geofence.color}20` }}
                          >
                            {geofence.icon || '🏢'}
                          </div>
                          <span className="font-medium text-gray-900 text-sm">{geofence.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">
                        {geofence.address || '-'}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-500 font-mono">
                        {geofence.latitude.toFixed(4)}, {geofence.longitude.toFixed(4)}
                      </td>
                      <td className="px-4 py-3">
                        <span 
                          className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium" 
                          style={{ backgroundColor: `${geofence.color}20`, color: geofence.color }}
                        >
                          {geofence.radius}m
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          geofence.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                        }`}>
                          ● {geofence.is_active ? 'Aktif' : 'Tidak Aktif'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleGeofenceSelect(geofence)}
                            className="text-blue-600 hover:text-blue-700 text-sm font-medium"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => {
                              if (confirm(`Hapus lokasi "${geofence.name}"?`)) {
                                geofenceApi.delete(geofence.id).then(() => {
                                  toast.success('Lokasi berhasil dihapus');
                                  loadGeofences();
                                });
                              }
                            }}
                            className="text-red-600 hover:text-red-700 text-sm font-medium"
                          >
                            Hapus
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
