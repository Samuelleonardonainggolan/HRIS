// lib/models/geofence_model.dart

class GeofenceModel {
  final String id;
  final String name;
  final double latitude;
  final double longitude;
  final int radius;
  final String address;
  final bool isActive;

  GeofenceModel({
    required this.id,
    required this.name,
    required this.latitude,
    required this.longitude,
    required this.radius,
    required this.address,
    required this.isActive,
  });

  factory GeofenceModel.fromJson(Map<String, dynamic> json) {
    return GeofenceModel(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      radius: (json['radius'] as num?)?.toInt() ?? 100,
      address: json['address']?.toString() ?? '',
      isActive: json['is_active'] == true,
    );
  }
}
