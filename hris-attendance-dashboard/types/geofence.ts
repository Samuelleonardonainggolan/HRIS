// src/types/geofence.ts

export interface Geofence {
  id: string;
  name: string;
  description: string;
  latitude: number;
  longitude: number;
  address: string;
  radius: number;
  radius_unit: string;
  icon: string;
  color: string;
  applies_to: 'all' | 'departments' | 'positions';
  department_ids?: string[];
  position_ids?: string[];
  is_active: boolean;
  created_by: string;
  created_by_name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGeofenceRequest {
  name: string;
  description?: string;
  latitude: number;
  longitude: number;
  address?: string;
  radius: number;
  icon?: string;
  color?: string;
  applies_to: 'all' | 'departments' | 'positions';
  department_ids?: string[];
  position_ids?: string[];
}

export interface UpdateGeofenceRequest {
  name?: string;
  description?: string;
  latitude?: number;
  longitude?: number;
  address?: string;
  radius?: number;
  icon?: string;
  color?: string;
  applies_to?: string;
  department_ids?: string[];
  position_ids?: string[];
  is_active?: boolean;
}

export interface CheckLocationResponse {
  is_within_geofence: boolean;
  geofence?: Geofence;
  distance: number;
  message: string;
}