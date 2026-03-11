'use client';

import { Geofence } from '@/types/geofence';
import { MapPinIcon, BuildingOfficeIcon } from '@heroicons/react/24/outline';

interface LocationListProps {
  geofences: Geofence[];
  selectedGeofence?: Geofence | null;
  onSelect: (geofence: Geofence) => void;
  onEdit: (geofence: Geofence) => void;
  onDelete: (geofence: Geofence) => void;
}

export default function LocationList({
  geofences,
  selectedGeofence,
  onSelect,
  onEdit,
  onDelete,
}: LocationListProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      <div className="p-4 border-b border-gray-200">
        <h3 className="text-lg font-semibold">Lokasi Terdaftar</h3>
        <p className="text-sm text-gray-500 mt-1">
          Daftar semua koordinat geofencing yang aktif
        </p>
      </div>

      <div className="divide-y divide-gray-200">
        {geofences.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <MapPinIcon className="w-12 h-12 mx-auto mb-2 text-gray-400" />
            <p>Belum ada lokasi terdaftar</p>
          </div>
        ) : (
          geofences.map((geofence) => (
            <div
              key={geofence.id}
              className={`p-4 cursor-pointer hover:bg-gray-50 transition-colors ${
                selectedGeofence?.id === geofence.id ? 'bg-blue-50' : ''
              }`}
              onClick={() => onSelect(geofence)}
            >
              <div className="flex items-start gap-3">
                {/* Icon */}
                <div
                  className="flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center text-2xl"
                  style={{ backgroundColor: `${geofence.color}20` }}
                >
                  {geofence.icon || '📍'}
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-1">
                    <h4 className="font-semibold text-gray-900 truncate">
                      {geofence.name}
                    </h4>
                    <span
                      className={`px-2 py-0.5 text-xs font-medium rounded-full ${
                        geofence.is_active
                          ? 'bg-green-100 text-green-700'
                          : 'bg-red-100 text-red-700'
                      }`}
                    >
                      {geofence.is_active ? '✓ Aktif' : '✗ Tidak Aktif'}
                    </span>
                  </div>

                  <p className="text-sm text-gray-600 mb-2 line-clamp-1">
                    {geofence.address || 'No address'}
                  </p>

                  <div className="flex items-center gap-4 text-xs text-gray-500">
                    <span className="flex items-center gap-1">
                      <MapPinIcon className="w-4 h-4" />
                      {geofence.latitude.toFixed(4)}, {geofence.longitude.toFixed(4)}
                    </span>
                    <span
                      className="px-2 py-0.5 rounded-full font-medium"
                      style={{ backgroundColor: `${geofence.color}20`, color: geofence.color }}
                    >
                      {geofence.radius}m
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onEdit(geofence);
                    }}
                    className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                    title="Edit"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      />
                    </svg>
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onDelete(geofence);
                    }}
                    className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                    title="Delete"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}