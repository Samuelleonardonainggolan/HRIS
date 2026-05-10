class User {
  final String id;
  final String nik;
  final String email;
  final String fullName;
  final String role;
  final String departmentId;
  final String department;
  final String position;
  final String? avatar;
  final String? phone;
  final String? address;
  final DateTime? birthDate;
  final String? religion;
  final String? lastEducation;
  final String? yearEnrolled;
  final String? employmentStatus;
  final DateTime joinDate;
  final bool isActive;

  User({
    required this.id,
    required this.nik,
    required this.email,
    required this.fullName,
    required this.role,
    required this.departmentId,
    required this.department,
    required this.position,
    this.avatar,
    this.phone,
    this.address,
    this.birthDate,
    this.religion,
    this.lastEducation,
    this.yearEnrolled,
    this.employmentStatus,
    required this.joinDate,
    required this.isActive,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    DateTime? birthDt;
    try {
      final rawBirth = json['birth_date'];
      if (rawBirth != null && rawBirth.toString().isNotEmpty) {
        birthDt = DateTime.parse(rawBirth.toString());
      }
    } catch (_) {
      birthDt = null;
    }

    DateTime joinDt;
    try {
      final rawJoin = json['created_at'] ?? json['join_date'];
      joinDt = rawJoin != null && rawJoin.toString().isNotEmpty
          ? DateTime.parse(rawJoin.toString())
          : DateTime.now();
    } catch (_) {
      joinDt = DateTime.now();
    }

    return User(
      id:               (json['id'] ?? json['_id'] ?? '').toString(),
      nik:              (json['payroll_number'] ?? json['nik'] ?? '').toString(),
      email:            (json['email'] ?? '').toString(),
      fullName:         (json['full_name'] ?? '').toString(),
      role:             (json['role'] ?? 'staf').toString(),
      departmentId:     (json['department_id'] ?? '').toString(),
      department:       (json['department_name'] ?? json['department'] ?? '').toString(),
      position:         (json['position_name']  ?? json['position']  ?? '').toString(),
      avatar:           json['avatar']?.toString(),
      phone:            json['phone']?.toString(),
      address:          json['address']?.toString(),
      birthDate:        birthDt,
      religion:         json['religion']?.toString(),
      lastEducation:    json['last_education']?.toString(),
      yearEnrolled:     json['year_enrolled']?.toString(),
      employmentStatus: json['employment_status']?.toString(),
      joinDate:         joinDt,
      isActive:         json['is_active'] == true || json['isActive'] == true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'nik': nik,
      'email': email,
      'full_name': fullName,
      'role': role,
      'department_id': departmentId,
      'department': department,
      'position': position,
      'avatar': avatar,
      'phone': phone,
      'address': address,
      'join_date': joinDate.toIso8601String(),
      'is_active': isActive,
    };
  }

    // Role check helper  s
    String get _normalizedRole => role.trim().toLowerCase().replaceAll(' ', '_');

    bool get isManagerHR =>
      _normalizedRole == 'manager_hr' ||
      _normalizedRole == 'hr_manager' ||
      _normalizedRole == 'managerhr';

    bool get isManagerDept =>
      _normalizedRole == 'manager_departemen' ||
      _normalizedRole == 'manager_department' ||
      _normalizedRole == 'kadep' ||
      (_normalizedRole.contains('manager') &&
        (_normalizedRole.contains('departemen') ||
          _normalizedRole.contains('department')));

    bool get isAdminDept =>
      _normalizedRole == 'admin_departemen' ||
      _normalizedRole == 'admin_department';

    bool get isStaf => _normalizedRole == 'staf' || _normalizedRole == 'staff';
  
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