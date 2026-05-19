// lib/pages/profile_page.dart
import 'dart:async';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:intl/intl.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mobile_app/services/sse_service.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});
  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  bool _isEditing = false;
  bool _isLoading = true;
  bool _isSaving = false;
  File? _profileImage;
  User? _user;
  StreamSubscription? _sseSubscription;

  final _phoneCtrl = TextEditingController();
  final _addressCtrl = TextEditingController();
  final _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  void initState() {
    super.initState();
    initializeDateFormatting('id', null);
    _setupSSE();
    _loadProfile();
  }

  void _setupSSE() {
    _sseSubscription = SSEService().events.listen((event) {
      if (!mounted || event.type == 'ping') return;
      _loadProfile(silent: true);
    });
  }

  Future<void> _loadProfile({bool silent = false}) async {
    if (!silent) {
      setState(() => _isLoading = true);
    }
    try {
      final u = await ApiService.getProfile();
      if (mounted) {
        setState(() {
          _user = u;
          _phoneCtrl.text = u.phone ?? '';
          _addressCtrl.text = u.address ?? '';
          _profileImage = null;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _saveChanges() async {
    setState(() => _isSaving = true);
    try {
      await ApiService.updateProfile({
        'phone': _phoneCtrl.text.trim(),
        'address': _addressCtrl.text.trim(),
      }, avatarPath: _profileImage?.path);
      await _loadProfile();
      if (mounted) {
        setState(() {
          _isEditing = false;
          _isSaving = false;
        });
        _snack('Profil berhasil diperbarui', AppTheme.successColor);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isSaving = false);
        _snack('Gagal menyimpan: $e', AppTheme.errorColor);
      }
    }
  }

  Future<void> _pickImage() async {
    final p = await ImagePicker().pickImage(source: ImageSource.gallery);
    if (p != null) setState(() => _profileImage = File(p.path));
  }

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 12) return 'Selamat Pagi';
    if (h < 15) return 'Selamat Siang';
    if (h < 18) return 'Selamat Sore';
    return 'Selamat Malam';
  }

  String _fmtDate(DateTime? dt) {
    if (dt == null) return '-';
    try {
      return DateFormat('dd MMMM yyyy', 'id').format(dt);
    } catch (_) {
      return '-';
    }
  }

  void _snack(String msg, Color color) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg, style: const TextStyle(fontWeight: FontWeight.w500)),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  String _avatarUrl() {
    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) {
      return avatar;
    }
    final n = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$n&background=135BEC&color=fff&size=100';
  }

  ImageProvider? _avatarImageProvider() {
    if (_profileImage != null) {
      return FileImage(_profileImage!);
    }

    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) {
      return NetworkImage(avatar);
    }

    final fallback = _avatarUrl().trim();
    if (fallback.isNotEmpty) {
      return NetworkImage(fallback);
    }

    return null;
  }

  void _showAvatarDetail() {
    final provider = _avatarImageProvider();
    if (provider == null) return;
    double totalVerticalDrag = 0;

    showDialog(
      context: context,
      barrierColor: Colors.black,
      builder: (_) {
        return Dialog.fullscreen(
          backgroundColor: Colors.black,
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onVerticalDragUpdate: (details) {
              if (details.delta.dy > 0) {
                totalVerticalDrag += details.delta.dy;
              }
            },
            onVerticalDragEnd: (details) {
              final shouldCloseByDistance = totalVerticalDrag > 120;
              final shouldCloseByVelocity =
                  (details.primaryVelocity ?? 0) > 850;
              totalVerticalDrag = 0;
              if (shouldCloseByDistance || shouldCloseByVelocity) {
                Navigator.of(context).pop();
              }
            },
            child: Stack(
              children: [
                Center(
                  child: InteractiveViewer(
                    minScale: 1,
                    maxScale: 4,
                    child: Image(
                      image: provider,
                      fit: BoxFit.contain,
                      errorBuilder: (_, __, ___) => const Icon(
                        Icons.broken_image_outlined,
                        color: Colors.white70,
                        size: 64,
                      ),
                    ),
                  ),
                ),
                Positioned(
                  top: 12,
                  right: 8,
                  child: SafeArea(
                    child: IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close, color: Colors.white),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _avatarPreview({double size = 90}) {
    if (_profileImage != null) {
      return Image.file(_profileImage!, fit: BoxFit.cover);
    }

    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) {
      return Image.network(
        avatar,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) =>
            const Icon(Icons.person, color: Color(0xFF135BEC), size: 44),
      );
    }

    return Image.network(
      _avatarUrl(),
      fit: BoxFit.cover,
      errorBuilder: (_, __, ___) =>
          const Icon(Icons.person, color: Color(0xFF135BEC), size: 44),
    );
  }

  @override
  void dispose() {
    _sseSubscription?.cancel();
    _phoneCtrl.dispose();
    _addressCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: _scaffoldKey,
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            _buildHeader(),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : RefreshIndicator(
                      onRefresh: _loadProfile,
                      child: SingleChildScrollView(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.fromLTRB(16, 20, 16, 16),
                        child: Column(
                          children: [
                            _buildProfileCard(),
                            const SizedBox(height: 16),
                            _buildPersonalInfo(),
                            const SizedBox(height: 16),
                            _buildEmploymentInfo(),
                            const SizedBox(height: 16),
                            _buildSettings(),
                            const SizedBox(height: 32),
                          ],
                        ),
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  // ── Header sama persis seperti profile ────────────────────────────────────
  Widget _buildHeader() {
    String displayRole = 'Karyawan';
    if (_user?.role != null && _user!.role.isNotEmpty) {
      final parts = _user!.role.split('_');
      displayRole = parts.map((w) => w.isNotEmpty ? '${w[0].toUpperCase()}${w.substring(1)}' : '').join(' ');
    }

    return Container(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 10),
      color: Colors.transparent,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          const Text(
            'Profil Saya',
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: Color(0xFF0F172A),
              letterSpacing: -0.5,
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: const Color(0xFF135BEC).withOpacity(0.1),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  Icons.verified_user_rounded,
                  color: Color(0xFF135BEC),
                  size: 14,
                ),
                const SizedBox(width: 6),
                Text(
                  displayRole,
                  style: const TextStyle(
                    color: Color(0xFF135BEC),
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Profile Hero Card — centered, full-width ──────────────────────────────
  Widget _buildProfileCard() {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF135BEC), Color(0xFF1E40AF), Color(0xFF1D4ED8)],
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.4),
            blurRadius: 24,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Stack(
        children: [
          // Decorative circles background
          Positioned(
            top: -30,
            right: -30,
            child: Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.white.withOpacity(0.06),
              ),
            ),
          ),
          Positioned(
            bottom: -20,
            left: -20,
            child: Container(
              width: 90,
              height: 90,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.white.withOpacity(0.05),
              ),
            ),
          ),
          // Content
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 20),
            child: Column(
              children: [
                // Avatar
                GestureDetector(
                  onTap: _showAvatarDetail,
                  child: Stack(
                    alignment: Alignment.bottomRight,
                    children: [
                      Container(
                        height: 90,
                        width: 90,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white, width: 3),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.2),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: ClipOval(child: _avatarPreview()),
                      ),
                      if (_isEditing)
                        GestureDetector(
                          onTap: _pickImage,
                          child: Container(
                            height: 26,
                            width: 26,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: Colors.white,
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.15),
                                  blurRadius: 6,
                                ),
                              ],
                            ),
                            child: const Icon(
                              Icons.camera_alt,
                              size: 14,
                              color: Color(0xFF135BEC),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 14),
                // Name
                Text(
                  _user?.fullName ?? '-',
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: 0.3,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 4),
                // Position
                Text(
                  _user?.position ?? '-',
                  style: TextStyle(
                    fontSize: 13,
                    color: Colors.white.withOpacity(0.8),
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 12),
                // Department badge + Status badge
                Wrap(
                  alignment: WrapAlignment.center,
                  spacing: 8,
                  children: [
                    _chip(Icons.business_outlined, _user?.department ?? '-'),
                    _chip(
                      Icons.circle,
                      (_user?.isActive ?? false) ? 'Aktif' : 'Non-Aktif',
                      color: (_user?.isActive ?? false)
                          ? const Color(0xFF2ECC71)
                          : const Color(0xFFEF4444),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                // Divider
                Divider(color: Colors.white.withOpacity(0.15), height: 1),
                const SizedBox(height: 16),
                // Stats row: NIK | Join date
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: [
                    _statItem('Payroll', _user?.nik ?? '-'),
                    _verticalDivider(),
                    _statItem('Bergabung', _fmtDateShort(_user?.joinDate)),
                  ],
                ),
                const SizedBox(height: 16),
                // Edit / Save button
                GestureDetector(
                  onTap: _isSaving
                      ? null
                      : () {
                          if (_isEditing)
                            _saveChanges();
                          else
                            setState(() => _isEditing = true);
                        },
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(14),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.1),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: _isSaving
                        ? const Center(
                            child: SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(
                                strokeWidth: 2.5,
                                color: Color(0xFF135BEC),
                              ),
                            ),
                          )
                        : Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                _isEditing
                                    ? Icons.check_circle_outline
                                    : Icons.edit_outlined,
                                color: const Color(0xFF135BEC),
                                size: 17,
                              ),
                              const SizedBox(width: 6),
                              Text(
                                _isEditing ? 'Simpan Perubahan' : 'Edit Profil',
                                style: const TextStyle(
                                  color: Color(0xFF135BEC),
                                  fontSize: 14,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                  ),
                ),
                if (_isEditing) ...[
                  const SizedBox(height: 8),
                  GestureDetector(
                    onTap: () => setState(() {
                      _isEditing = false;
                      _phoneCtrl.text = _user?.phone ?? '';
                      _addressCtrl.text = _user?.address ?? '';
                      _profileImage = null;
                    }),
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.15),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(
                          color: Colors.white.withOpacity(0.4),
                        ),
                      ),
                      child: const Center(
                        child: Text(
                          'Batal',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 13,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _chip(IconData icon, String label, {Color? color}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.15),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withOpacity(0.25)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 10, color: color ?? Colors.white),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              fontSize: 11,
              color: Colors.white,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _statItem(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
          textAlign: TextAlign.center,
          overflow: TextOverflow.ellipsis,
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: TextStyle(fontSize: 10, color: Colors.white.withOpacity(0.7)),
        ),
      ],
    );
  }

  Widget _verticalDivider() {
    return Container(
      height: 28,
      width: 1,
      color: Colors.white.withOpacity(0.2),
    );
  }

  String _fmtDateShort(DateTime? dt) {
    if (dt == null) return '-';
    try {
      return DateFormat('MMM yyyy', 'id').format(dt);
    } catch (_) {
      return '-';
    }
  }

  Widget _buildPersonalInfo() {
    return _card(
      'Informasi Pribadi',
      Icons.person_outline,
      const Color(0xFF6366F1),
      [
        _row(Icons.email_outlined, 'Email', _user?.email ?? '-'),
        _div(),
        _row(
          Icons.phone_outlined,
          'Nomor HP',
          _user?.phone ?? '-',
          editable: _isEditing,
          ctrl: _phoneCtrl,
        ),
        _div(),
        _row(Icons.cake_outlined, 'Tanggal Lahir', _fmtDate(_user?.birthDate)),
        _div(),
        _row(Icons.favorite_outline, 'Agama', _user?.religion ?? '-'),
        _div(),
        _row(
          Icons.school_outlined,
          'Pendidikan Terakhir',
          _user?.lastEducation ?? '-',
        ),
        _div(),
        _row(
          Icons.calendar_month_outlined,
          'Tahun Masuk',
          _user?.yearEnrolled ?? '-',
        ),
        _div(),
        _row(
          Icons.work_outline,
          'Status Kerja',
          _user?.employmentStatus ?? '-',
        ),
        _div(),
        _row(
          Icons.calendar_today_outlined,
          'Bergabung',
          _fmtDate(_user?.joinDate),
        ),
        _div(),
        _row(
          Icons.location_on_outlined,
          'Alamat',
          _user?.address ?? '-',
          editable: _isEditing,
          ctrl: _addressCtrl,
          multi: true,
        ),
      ],
    );
  }

  Widget _buildEmploymentInfo() {
    return _card(
      'Detail Kepegawaian',
      Icons.work_outline,
      const Color(0xFF0EA5E9),
      [
        _row(Icons.badge_outlined, 'NIK / Payroll', _user?.nik ?? '-'),
        _div(),
        _row(Icons.assignment_ind_outlined, 'Jabatan', _user?.position ?? '-'),
        _div(),
        _row(Icons.business_outlined, 'Departemen', _user?.department ?? '-'),
        _div(),
        _row(
          Icons.toggle_on_outlined,
          'Status',
          (_user?.isActive ?? false) ? 'Aktif' : 'Non-Aktif',
        ),
      ],
    );
  }

  Widget _buildSettings() {
    return _card(
      'Pengaturan',
      Icons.settings_outlined,
      const Color(0xFF64748B),
      [
        _settingTile(
          Icons.lock_outlined,
          'Ganti Password',
          onTap: _showChangePasswordDialog,
        ),
      ],
    );
  }

  Widget _card(
    String title,
    IconData icon,
    Color accentColor,
    List<Widget> children,
  ) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Card header with accent color bar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
            decoration: BoxDecoration(
              color: accentColor.withOpacity(0.06),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(20),
              ),
              border: Border(
                bottom: BorderSide(color: accentColor.withOpacity(0.1)),
              ),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: accentColor.withOpacity(0.12),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(icon, size: 17, color: accentColor),
                ),
                const SizedBox(width: 10),
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: accentColor,
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 8),
            child: Column(children: children),
          ),
        ],
      ),
    );
  }

  Widget _row(
    IconData icon,
    String label,
    String value, {
    bool editable = false,
    TextEditingController? ctrl,
    bool multi = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 9),
      child: Row(
        crossAxisAlignment: (editable && _isEditing) || multi
            ? CrossAxisAlignment.start
            : CrossAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(7),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, size: 15, color: const Color(0xFF475569)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade500,
                    fontWeight: FontWeight.w600,
                    letterSpacing: 0.3,
                  ),
                ),
                const SizedBox(height: 3),
                if (editable && _isEditing && ctrl != null)
                  TextField(
                    controller: ctrl,
                    maxLines: multi ? 3 : 1,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF0F172A),
                    ),
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 8,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(
                          color: Color(0xFF135BEC),
                          width: 1.5,
                        ),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(
                          color: Color(0xFF135BEC),
                          width: 1.5,
                        ),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: Colors.grey.shade300),
                      ),
                      hintText: 'Masukkan $label',
                      hintStyle: TextStyle(
                        color: Colors.grey.shade400,
                        fontSize: 13,
                      ),
                      filled: true,
                      fillColor: const Color(0xFFF8FAFC),
                    ),
                  )
                else
                  Text(
                    value,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF0F172A),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _div() =>
      Divider(color: const Color(0xFFF1F5F9), height: 1, thickness: 1);

  Widget _settingTile(
    IconData icon,
    String title, {
    Widget? trailing,
    VoidCallback? onTap,
    Color? iconColor,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 10),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(9),
              decoration: BoxDecoration(
                color: (iconColor ?? const Color(0xFF135BEC)).withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                icon,
                size: 17,
                color: iconColor ?? const Color(0xFF135BEC),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                title,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: Color(0xFF0F172A),
                ),
              ),
            ),
            trailing ??
                Container(
                  padding: const EdgeInsets.all(4),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: const Icon(
                    Icons.chevron_right,
                    color: Color(0xFF94A3B8),
                    size: 16,
                  ),
                ),
          ],
        ),
      ),
    );
  }

  // ── Ganti Password — full screen bottom sheet modern ───────────────────────
  void _showChangePasswordDialog() {
    final oldCtrl = TextEditingController();
    final newCtrl = TextEditingController();
    final confCtrl = TextEditingController();
    bool showOld = false;
    bool showNew = false;
    bool showConf = false;
    bool hasMin8 = false;
    bool hasUpper = false;
    bool hasNumber = false;
    bool hasSpecial = false;

    void checkStrength(String val) {
      hasMin8 = val.length >= 8;
      hasUpper = val.contains(RegExp(r'[A-Z]'));
      hasNumber = val.contains(RegExp(r'[0-9]'));
      hasSpecial = val.contains(RegExp(r'[!@#\$%^&*(),.?":{}|<>]'));
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setSheet) => Container(
          padding: EdgeInsets.only(
            left: 24,
            right: 24,
            top: 24,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Handle bar
                Center(
                  child: Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                const SizedBox(height: 20),
                // Title
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: const Color(0xFF135BEC).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(
                        Icons.lock_reset_rounded,
                        color: Color(0xFF135BEC),
                        size: 22,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Ganti Password',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF0F172A),
                          ),
                        ),
                        Text(
                          'Pastikan password kuat & aman',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade500,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                // Current password
                _pwFieldSheet(
                  oldCtrl,
                  'Password Saat Ini',
                  Icons.lock_outline,
                  showOld,
                  () => setSheet(() => showOld = !showOld),
                ),
                const SizedBox(height: 16),
                // New password with live check
                _pwFieldSheet(
                  newCtrl,
                  'Password Baru',
                  Icons.lock_open_rounded,
                  showNew,
                  () => setSheet(() => showNew = !showNew),
                  onChange: (v) => setSheet(() => checkStrength(v)),
                ),
                const SizedBox(height: 12),
                // Strength indicators
                Container(
                  padding: const EdgeInsets.all(14),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Syarat Password:',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: Colors.grey.shade600,
                        ),
                      ),
                      const SizedBox(height: 8),
                      _req(hasMin8, 'Minimal 8 karakter'),
                      _req(hasUpper, 'Mengandung huruf kapital (A-Z)'),
                      _req(hasNumber, 'Mengandung angka (0-9)'),
                      _req(hasSpecial, 'Mengandung karakter khusus (!@#\$...)'),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                // Confirm password
                _pwFieldSheet(
                  confCtrl,
                  'Konfirmasi Password Baru',
                  Icons.check_circle_outline,
                  showConf,
                  () => setSheet(() => showConf = !showConf),
                ),
                const SizedBox(height: 24),
                // Submit button
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF135BEC),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(14),
                      ),
                      elevation: 0,
                    ),
                    onPressed: () async {
                      if (oldCtrl.text.isEmpty ||
                          newCtrl.text.isEmpty ||
                          confCtrl.text.isEmpty) {
                        _snack('Semua field wajib diisi', AppTheme.errorColor);
                        return;
                      }
                      if (!hasMin8) {
                        _snack(
                          'Password minimal 8 karakter',
                          AppTheme.errorColor,
                        );
                        return;
                      }
                      if (!hasUpper) {
                        _snack('Harus ada huruf kapital', AppTheme.errorColor);
                        return;
                      }
                      if (!hasNumber) {
                        _snack('Harus ada angka', AppTheme.errorColor);
                        return;
                      }
                      if (!hasSpecial) {
                        _snack(
                          'Harus ada karakter khusus',
                          AppTheme.errorColor,
                        );
                        return;
                      }
                      if (newCtrl.text != confCtrl.text) {
                        _snack('Konfirmasi tidak cocok', AppTheme.errorColor);
                        return;
                      }
                      Navigator.pop(ctx);
                      try {
                        await ApiService.changePassword(
                          oldPassword: oldCtrl.text,
                          newPassword: newCtrl.text,
                        );
                        _snack(
                          'Password berhasil diubah',
                          AppTheme.successColor,
                        );
                      } catch (e) {
                        _snack('Gagal: $e', AppTheme.errorColor);
                      }
                    },
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.check_rounded, size: 18),
                        SizedBox(width: 8),
                        Text(
                          'Ubah Password',
                          style: TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                SizedBox(
                  width: double.infinity,
                  child: TextButton(
                    onPressed: () => Navigator.pop(ctx),
                    child: Text(
                      'Batal',
                      style: TextStyle(
                        color: Colors.grey.shade500,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _pwFieldSheet(
    TextEditingController c,
    String label,
    IconData icon,
    bool show,
    VoidCallback toggle, {
    ValueChanged<String>? onChange,
  }) {
    return TextField(
      controller: c,
      obscureText: !show,
      onChanged: onChange,
      decoration: InputDecoration(
        labelText: label,
        labelStyle: TextStyle(fontSize: 13, color: Colors.grey.shade500),
        prefixIcon: Icon(icon, size: 18, color: const Color(0xFF135BEC)),
        suffixIcon: IconButton(
          icon: Icon(
            show ? Icons.visibility_off_outlined : Icons.visibility_outlined,
            size: 18,
            color: Colors.grey.shade400,
          ),
          onPressed: toggle,
        ),
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey.shade200),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF135BEC), width: 1.5),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 14,
        ),
      ),
    );
  }

  Widget _req(bool ok, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: [
          Icon(
            ok
                ? Icons.check_circle_rounded
                : Icons.radio_button_unchecked_rounded,
            size: 14,
            color: ok ? const Color(0xFF2ECC71) : Colors.grey.shade400,
          ),
          const SizedBox(width: 6),
          Text(
            text,
            style: TextStyle(
              fontSize: 12,
              color: ok ? const Color(0xFF2ECC71) : Colors.grey.shade500,
              fontWeight: ok ? FontWeight.w500 : FontWeight.normal,
            ),
          ),
        ],
      ),
    );
  }
}
