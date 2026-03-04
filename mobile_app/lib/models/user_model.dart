class User {
  final String id;
  final String nik;
  final String email;
  final String fullName;
  final String role;
  final String department;
  final String position;
  final String? avatar;
  final String? phone;
  final String? address;
  final DateTime joinDate;
  final bool isActive;

  User({
    required this.id,
    required this.nik,
    required this.email,
    required this.fullName,
    required this.role,
    required this.department,
    required this.position,
    this.avatar,
    this.phone,
    this.address,
    required this.joinDate,
    required this.isActive,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? json['_id'] ?? '',
      nik: json['nik'] ?? '',
      email: json['email'] ?? '',
      fullName: json['full_name'] ?? '',
      role: json['role'] ?? 'staf',
      department: json['department'] ?? '',
      position: json['position'] ?? '',
      avatar: json['avatar'],
      phone: json['phone'],
      address: json['address'],
      joinDate: DateTime.parse(json['join_date'] ?? DateTime.now().toIso8601String()),
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'nik': nik,
      'email': email,
      'full_name': fullName,
      'role': role,
      'department': department,
      'position': position,
      'avatar': avatar,
      'phone': phone,
      'address': address,
      'join_date': joinDate.toIso8601String(),
      'is_active': isActive,
    };
  }

  // Role check helpers
  bool get isManagerHR => role == 'manager_hr';
  bool get isManagerDept => role == 'manager_departemen';
  bool get isAdminDept => role == 'admin_departemen';
  bool get isStaf => role == 'staf';
  
  bool hasPermission(String permission) {
    final permissions = {
      'manager_hr': [
        'user:create', 'user:read', 'user:update', 'user:delete',
        'attendance:approve', 'attendance:view_all',
        'department:manage', 'report:view_all',
      ],
      'manager_departemen': [
        'user:read', 'user:update_dept',
        'attendance:approve_dept', 'attendance:view_dept',
        'report:view_dept',
      ],
      'admin_departemen': [
        'user:read', 'user:create_staf', 'user:update_staf',
        'attendance:view_dept', 'attendance:input',
      ],
      'staf': [
        'attendance:submit', 'profile:view', 'profile:update',
      ],
    };

    final rolePermissions = permissions[role];
    if (rolePermissions == null) return false;
    
    return rolePermissions.contains(permission);
  }
}

class UpdateProfileRequest {
  final String? fullName;
  final String? phone;
  final String? address;
  final String? avatar;
  final String? department;
  final String? position;

  UpdateProfileRequest({
    this.fullName,
    this.phone,
    this.address,
    this.avatar,
    this.department,
    this.position,
  });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = {};
    if (fullName != null) data['full_name'] = fullName;
    if (phone != null) data['phone'] = phone;
    if (address != null) data['address'] = address;
    if (avatar != null) data['avatar'] = avatar;
    if (department != null) data['department'] = department;
    if (position != null) data['position'] = position;
    return data;
  }
}

class ChangePasswordRequest {
  final String oldPassword;
  final String newPassword;

  ChangePasswordRequest({
    required this.oldPassword,
    required this.newPassword,
  });

  Map<String, dynamic> toJson() {
    return {
      'old_password': oldPassword,
      'new_password': newPassword,
    };
  }
}