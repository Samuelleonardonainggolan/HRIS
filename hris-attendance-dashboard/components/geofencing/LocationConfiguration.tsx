'use client';

import { useState } from 'react';
import { Geofence, CreateGeofenceRequest } from '@/types/geofence';

interface LocationConfigurationProps {
  selectedGeofence?: Geofence | null;
  onSave: (data: CreateGeofenceRequest) => void;
  onCancel: () => void;
  latitude?: number;
  longitude?: number;
}

export default function LocationConfiguration({
  selectedGeofence,
  onSave,
  onCancel,
  latitude,
  longitude,
}: LocationConfigurationProps) {
  const key =
    selectedGeofence?.id ??
    `${latitude ?? ''}:${longitude ?? ''}`;

  return (
    <LocationConfigurationInner
      key={key}
      selectedGeofence={selectedGeofence}
      onSave={onSave}
      onCancel={onCancel}
      latitude={latitude}
      longitude={longitude}
    />
  );
}

function LocationConfigurationInner({
  selectedGeofence,
  onSave,
  onCancel,
  latitude,
  longitude,
}: LocationConfigurationProps) {
  const [formData, setFormData] = useState<CreateGeofenceRequest>(() => {
    if (selectedGeofence) {
      return {
        name: selectedGeofence.name,
        description: selectedGeofence.description,
        latitude: selectedGeofence.latitude,
        longitude: selectedGeofence.longitude,
        address: selectedGeofence.address,
        radius: selectedGeofence.radius,
        icon: selectedGeofence.icon,
        color: selectedGeofence.color,
        applies_to: selectedGeofence.applies_to,
        department_ids: selectedGeofence.department_ids,
        position_ids: selectedGeofence.position_ids,
      };
    }

    return {
      name: '',
      description: '',
      latitude: latitude || 0,
      longitude: longitude || 0,
      address: '',
      radius: 150,
      icon: '📍',
      color: '#3B82F6',
      applies_to: 'all',
    };
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const icons = ['🏨', '🏢', '🍽️', '💼', '📍', '🏪', '🏭', '🏦', '🏛️'];
  const colors = [
    '#3B82F6', // Blue
    '#10B981', // Green
    '#F59E0B', // Orange
    '#EF4444', // Red
    '#8B5CF6', // Purple
    '#EC4899', // Pink
    '#06B6D4', // Cyan
  ];

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Konfigurasi Lokasi
          </h3>
          <p className="text-sm text-gray-500 mt-1">
            {selectedGeofence ? 'Edit lokasi geofencing' : 'Tambah lokasi baru'}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nama Lokasi <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="e.g., Hotel Labosta"
            required
          />
        </div>

        {/* Coordinates */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Latitude <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              step="0.000001"
              value={formData.latitude}
              onChange={(e) => setFormData({ ...formData, latitude: parseFloat(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="-6.2088"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Longitude <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              step="0.000001"
              value={formData.longitude}
              onChange={(e) => setFormData({ ...formData, longitude: parseFloat(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="106.8456"
              required
            />
          </div>
        </div>

        {/* Radius */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Radius (Meter) <span className="text-red-500">*</span>
          </label>
          <div className="flex items-center gap-4">
            <input
              type="range"
              min="10"
              max="1000"
              step="10"
              value={formData.radius}
              onChange={(e) => setFormData({ ...formData, radius: parseInt(e.target.value) })}
              className="flex-1"
            />
            <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-lg font-semibold min-w-[80px] text-center">
              {formData.radius}m
            </span>
          </div>
          <p className="text-xs text-gray-500 mt-1">Range: 10m - 1000m</p>
        </div>

        {/* Icon Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Icon</label>
          <div className="flex gap-2 flex-wrap">
            {icons.map((icon) => (
              <button
                key={icon}
                type="button"
                onClick={() => setFormData({ ...formData, icon })}
                className={`w-12 h-12 rounded-lg border-2 flex items-center justify-center text-2xl hover:scale-110 transition-transform ${
                  formData.icon === icon ? 'border-blue-500 bg-blue-50' : 'border-gray-200'
                }`}
              >
                {icon}
              </button>
            ))}
          </div>
        </div>

        {/* Color Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Warna</label>
          <div className="flex gap-2">
            {colors.map((color) => (
              <button
                key={color}
                type="button"
                onClick={() => setFormData({ ...formData, color })}
                className={`w-10 h-10 rounded-lg border-2 hover:scale-110 transition-transform ${
                  formData.color === color ? 'border-gray-800' : 'border-gray-200'
                }`}
                style={{ backgroundColor: color }}
              />
            ))}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-3 pt-4">
          <button
            type="submit"
            className="flex-1 bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 transition-colors"
          >
            Simpan Perubahan
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-6 py-2.5 border border-gray-300 rounded-lg font-medium hover:bg-gray-50 transition-colors"
          >
            Reset ke Default
          </button>
        </div>
      </form>
    </div>
  );
}