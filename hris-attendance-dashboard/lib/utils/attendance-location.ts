import type { Geofence } from "@/types/geofence";

type AttendanceGeofence = Pick<Geofence, "name" | "latitude" | "longitude" | "radius" | "radius_unit">;

function parseLocationCoordinates(rawLocation: string): { latitude: number; longitude: number } | null {
  const matches = rawLocation.trim().match(/-?\d+(?:\.\d+)?/g);
  if (!matches || matches.length < 2) return null;

  const latitude = Number(matches[0]);
  const longitude = Number(matches[1]);

  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return null;
  if (Math.abs(latitude) > 90 || Math.abs(longitude) > 180) return null;

  return { latitude, longitude };
}

function radiusToMeters(geofence: AttendanceGeofence): number {
  const radius = Number(geofence.radius);
  if (!Number.isFinite(radius) || radius <= 0) return 0;

  const unit = geofence.radius_unit?.toLowerCase();
  if (unit === "km" || unit === "kilometer" || unit === "kilometers") {
    return radius * 1000;
  }

  return radius;
}

function distanceMeters(
  latitude1: number,
  longitude1: number,
  latitude2: number,
  longitude2: number
): number {
  const earthRadiusMeters = 6371000;
  const toRadians = (value: number) => (value * Math.PI) / 180;

  const deltaLatitude = toRadians(latitude2 - latitude1);
  const deltaLongitude = toRadians(longitude2 - longitude1);
  const startLatitude = toRadians(latitude1);
  const endLatitude = toRadians(latitude2);

  const a =
    Math.sin(deltaLatitude / 2) * Math.sin(deltaLatitude / 2) +
    Math.sin(deltaLongitude / 2) * Math.sin(deltaLongitude / 2) *
      Math.cos(startLatitude) *
      Math.cos(endLatitude);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

  return earthRadiusMeters * c;
}

export function resolveAttendanceLocationLabel(
  rawLocation: string | null | undefined,
  geofences: AttendanceGeofence[]
): string {
  const location = rawLocation?.trim();
  if (!location) return "Unrecorded";

  const coordinates = parseLocationCoordinates(location);
  if (!coordinates) return location;

  const matchedGeofence = geofences.find((geofence) => {
    const radiusMeters = radiusToMeters(geofence);
    if (radiusMeters <= 0) return false;

    return (
      distanceMeters(
        coordinates.latitude,
        coordinates.longitude,
        geofence.latitude,
        geofence.longitude
      ) <= radiusMeters
    );
  });

  return matchedGeofence?.name ?? "Lokasi tidak dikenali";
}
