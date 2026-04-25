'use client';

import { useCallback, useState, memo } from 'react';
import { GoogleMap, useJsApiLoader, Marker, Circle, InfoWindow } from '@react-google-maps/api';
import { Geofence } from '@/types/geofence';

const mapContainerStyle = {
  width: '100%',
  height: '100%',
  borderRadius: '0.5rem',
};

// ✅ FIX 1: Move libraries outside component to prevent re-creation
const libraries: ("places" | "geometry")[] = ['places', 'geometry'];

// ✅ FIX 2: Move mapOptions outside component
const mapOptions: google.maps.MapOptions = {
  mapTypeControl: true,
  fullscreenControl: true,
  streetViewControl: false,
  zoomControl: true,
};

interface GeofencingMapProps {
  geofences: Geofence[];
  selectedGeofence?: Geofence | null;
  onGeofenceSelect?: (geofence: Geofence) => void;
  onMapClick?: (lat: number, lng: number) => void;
  center?: { lat: number; lng: number };
  zoom?: number;
}

function GeofencingMap({
  geofences,
  selectedGeofence,
  onGeofenceSelect,
  onMapClick,
  center = { lat: -6.2088, lng: 106.8456 },
  zoom = 15,
}: GeofencingMapProps) {
  const [selectedMarker, setSelectedMarker] = useState<Geofence | null>(null);
  const [map, setMap] = useState<google.maps.Map | null>(null);

  // ✅ FIX 3: Get API key once
  const apiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || '';

  console.log('🗺️ GeofencingMap render - Geofences:', geofences.length);

  const { isLoaded, loadError } = useJsApiLoader({
    googleMapsApiKey: apiKey,
    libraries, // Now stable reference
  });

  // ✅ FIX 4: Properly memoize callbacks
  const onLoad = useCallback((mapInstance: google.maps.Map) => {
    console.log('✅ Map loaded successfully');
    setMap(mapInstance);
  }, []);

  const onUnmount = useCallback(() => {
    console.log('🗺️ Map unmounted');
    setMap(null);
  }, []);

  const handleMapClick = useCallback(
    (e: google.maps.MapMouseEvent) => {
      if (e.latLng && onMapClick) {
        const lat = e.latLng.lat();
        const lng = e.latLng.lng();
        console.log('🖱️ Map clicked:', { lat, lng });
        onMapClick(lat, lng);
        setSelectedMarker(null);
      }
    },
    [onMapClick]
  );

  const handleMarkerClick = useCallback(
    (geofence: Geofence) => {
      console.log('📍 Marker clicked:', geofence.name);
      setSelectedMarker(geofence);
      onGeofenceSelect?.(geofence);
    },
    [onGeofenceSelect]
  );

  const handleZoomIn = useCallback(() => {
    if (map) {
      const currentZoom = map.getZoom() || 15;
      map.setZoom(currentZoom + 1);
    }
  }, [map]);

  const handleZoomOut = useCallback(() => {
    if (map) {
      const currentZoom = map.getZoom() || 15;
      map.setZoom(currentZoom - 1);
    }
  }, [map]);

  const handleResetView = useCallback(() => {
    if (map) {
      map.setCenter(center);
      map.setZoom(zoom);
    }
  }, [map, center, zoom]);

  // Loading state
  if (!isLoaded) {
    console.log('⏳ Map loading...');
    return (
      <div className="relative w-full h-full flex items-center justify-center bg-gray-100 rounded-lg">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600 font-medium">Loading map...</p>
          <p className="text-gray-500 text-sm mt-2">Please wait</p>
        </div>
      </div>
    );
  }

  // Error state
  if (loadError) {
    console.error('❌ Map load error:', loadError);
    return (
      <div className="relative w-full h-full flex items-center justify-center bg-red-50 rounded-lg">
        <div className="text-center p-6 max-w-md">
          <svg className="w-12 h-12 text-red-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 className="text-lg font-semibold text-red-900 mb-2">Failed to load map</h3>
          <p className="text-sm text-red-700 mb-4">{loadError.message || 'Unknown error'}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  console.log('✅ Rendering map with', geofences.length, 'geofences');

  return (
    <div className="relative w-full h-full">
      <GoogleMap
        mapContainerStyle={mapContainerStyle}
        center={selectedGeofence ? { lat: selectedGeofence.latitude, lng: selectedGeofence.longitude } : center}
        zoom={selectedGeofence ? 17 : zoom}
        onClick={handleMapClick}
        options={mapOptions}
        onLoad={onLoad}
        onUnmount={onUnmount}
      >
        {/* Render Circles */}
        {geofences.map((geofence) => (
          <Circle
            key={`circle-${geofence.id}`}
            center={{ lat: geofence.latitude, lng: geofence.longitude }}
            radius={geofence.radius}
            options={{
              fillColor: geofence.color || '#3B82F6',
              fillOpacity: selectedGeofence?.id === geofence.id ? 0.35 : 0.2,
              strokeColor: geofence.color || '#3B82F6',
              strokeOpacity: 0.8,
              strokeWeight: selectedGeofence?.id === geofence.id ? 3 : 2,
              clickable: true,
            }}
            onClick={() => handleMarkerClick(geofence)}
          />
        ))}

        {/* Render Markers */}
        {geofences.map((geofence) => (
          <Marker
            key={`marker-${geofence.id}`}
            position={{ lat: geofence.latitude, lng: geofence.longitude }}
            title={geofence.name}
            label={{
              text: geofence.icon || '📍',
              fontSize: '24px',
            }}
            animation={selectedGeofence?.id === geofence.id ? google.maps.Animation.BOUNCE : undefined}
            onClick={() => handleMarkerClick(geofence)}
          />
        ))}

        {/* Render Info Window */}
        {selectedMarker && (
          <InfoWindow
            position={{ lat: selectedMarker.latitude, lng: selectedMarker.longitude }}
            onCloseClick={() => setSelectedMarker(null)}
          >
            <div className="p-2 min-w-[200px]">
              <h3 className="font-bold text-base mb-2 flex items-center gap-2">
                <span className="text-2xl">{selectedMarker.icon}</span>
                <span>{selectedMarker.name}</span>
              </h3>
              <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                {selectedMarker.address || 'No address available'}
              </p>
              <div className="flex items-center gap-3 text-xs text-gray-500 mb-2">
                <span className="flex items-center gap-1">
                  📏 {selectedMarker.radius}m
                </span>
                <span className="flex items-center gap-1">
                  📍 {selectedMarker.latitude.toFixed(4)}, {selectedMarker.longitude.toFixed(4)}
                </span>
              </div>
              <div className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${selectedMarker.is_active
                  ? 'bg-green-100 text-green-700'
                  : 'bg-red-100 text-red-700'
                }`}>
                {selectedMarker.is_active ? '✓ Active' : '✗ Inactive'}
              </div>
            </div>
          </InfoWindow>
        )}
      </GoogleMap>

      {/* Map Controls */}
      <div className="absolute top-4 right-4 flex flex-col gap-2 z-10">
        <button
          onClick={handleZoomIn}
          className="bg-white p-2 rounded-lg shadow-md hover:bg-gray-50 transition-colors"
          title="Zoom In"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
        <button
          onClick={handleZoomOut}
          className="bg-white p-2 rounded-lg shadow-md hover:bg-gray-50 transition-colors"
          title="Zoom Out"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
          </svg>
        </button>
        <button
          onClick={handleResetView}
          className="bg-white p-2 rounded-lg shadow-md hover:bg-gray-50 transition-colors"
          title="Reset View"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>

      {/* Legend */}
      <div className="absolute bottom-4 left-4 bg-white p-4 rounded-lg shadow-md z-10">
        <h4 className="font-semibold text-sm mb-2">Legend</h4>
        <div className="space-y-1.5 text-xs">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-blue-500"></div>
            <span>Geofence Area</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-green-500"></div>
            <span>Active</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500"></div>
            <span>Inactive</span>
          </div>
        </div>
      </div>

      {/* Geofence Count */}
      {geofences.length > 0 && (
        <div className="absolute top-4 left-4 bg-white px-3 py-2 rounded-lg shadow-md z-10">
          <p className="text-sm font-medium text-gray-700">
            📍 {geofences.length} {geofences.length === 1 ? 'Location' : 'Locations'}
          </p>
        </div>
      )}
    </div>
  );
}

// ✅ FIX 5: Memoize the component to prevent unnecessary re-renders
export default memo(GeofencingMap);