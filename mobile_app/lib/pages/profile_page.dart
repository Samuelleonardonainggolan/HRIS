// lib/pages/profile_page.dart
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:intl/intl.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:image_picker/image_picker.dart';

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

  final _phoneCtrl = TextEditingController();
  final _addressCtrl = TextEditingController();
  final _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  void initState() {
    super.initState();
    initializeDateFormatting('id', null);
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    setState(() => _isLoading = true);
    try {
      final u = await ApiService.getProfile();
      if (mounted) {
        setState(() {
          _user = u;
          _phoneCtrl.text = u.phone ?? '';
          _addressCtrl.text = u.address ?? '';
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
      });
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
    final n = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$n&background=135BEC&color=fff&size=100';
  }

  @override
  void dispose() {
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
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          children: [
                            _buildProfileCard(),
                            const SizedBox(height: 12),
                            _buildPersonalInfo(),
                            const SizedBox(height: 12),
                            _buildEmploymentInfo(),
                            const SizedBox(height: 12),
                            _buildSettings(),
                            const SizedBox(height: 24),
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
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(28)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Stack(
            children: [
              Hero(
                tag: 'profile',
                child: Container(
                  height: 48,
                  width: 48,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: const LinearGradient(
                      colors: [Color(0xFF135BEC), Color(0xFF3B7BF6)],
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF135BEC).withOpacity(0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(2),
                    child: Container(
                      decoration: const BoxDecoration(
                        shape: BoxShape.circle,
                        color: Colors.white,
                      ),
                      child: ClipOval(
                        child: _profileImage != null
                            ? Image.file(_profileImage!, fit: BoxFit.cover)
                            : Image.network(
                                _avatarUrl(),
                                fit: BoxFit.cover,
                                errorBuilder: (_, __, ___) => const Icon(
                                  Icons.person,
                                  color: Color(0xFF135BEC),
                                  size: 26,
                                ),
                              ),
                      ),
                    ),
                  ),
                ),
              ),
              Positioned(
                bottom: 1,
                right: 1,
                child: Container(
                  height: 12,
                  width: 12,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: const Color(0xFF2ECC71),
                    border: Border.all(color: Colors.white, width: 2),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _greeting(),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade500,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                Text(
                  _user?.fullName ?? 'Profil Saya',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF0F172A),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          Stack(
            children: [
              Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  shape: BoxShape.circle,
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.notifications_none,
                    color: Color(0xFF475569),
                    size: 22,
                  ),
                  onPressed: () {},
                  padding: EdgeInsets.zero,
                ),
              ),
              Positioned(
                top: 9,
                right: 9,
                child: Container(
                  height: 8,
                  width: 8,
                  decoration: const BoxDecoration(
                    shape: BoxShape.circle,
                    color: Color(0xFFEF4444),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // ── Profile Card — compact horizontal ─────────────────────────────────────
  Widget _buildProfileCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF135BEC), Color(0xFF2563EB)],
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.35),
            blurRadius: 16,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Row(
        children: [
          // Avatar kecil di card
          GestureDetector(
            onTap: _isEditing ? _pickImage : null,
            child: Stack(
              alignment: Alignment.bottomRight,
              children: [
                Container(
                  height: 64,
                  width: 64,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.white.withOpacity(0.8),
                      width: 2.5,
                    ),
                  ),
                  child: ClipOval(
                    child: _profileImage != null
                        ? Image.file(_profileImage!, fit: BoxFit.cover)
                        : Image.network(
                            _avatarUrl(),
                            fit: BoxFit.cover,
                            errorBuilder: (_, __, ___) => Container(
                              color: Colors.white,
                              child: const Icon(
                                Icons.person,
                                color: Color(0xFF135BEC),
                                size: 32,
                              ),
                            ),
                          ),
                  ),
                ),
                if (_isEditing)
                  Container(
                    height: 22,
                    width: 22,
                    decoration: const BoxDecoration(
                      shape: BoxShape.circle,
                      color: Colors.white,
                    ),
                    child: const Icon(
                      Icons.camera_alt,
                      size: 13,
                      color: Color(0xFF135BEC),
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 14),
          // Info nama, jabatan, dept
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _user?.fullName ?? '-',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  _user?.position ?? '-',
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.white.withOpacity(0.85),
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 3,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        _user?.department ?? '-',
                        style: const TextStyle(
                          fontSize: 11,
                          color: Colors.white,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          // Edit button compact
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
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.2),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.white.withOpacity(0.4)),
              ),
              child: _isSaving
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 2,
                      ),
                    )
                  : Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          _isEditing ? Icons.check : Icons.edit_outlined,
                          color: Colors.white,
                          size: 14,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          _isEditing ? 'Simpan' : 'Edit',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPersonalInfo() {
    return _card('Informasi Pribadi', Icons.person_outline, [
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
      _row(
        Icons.cake_outlined,
        'Tanggal Lahir',
        _fmtDate(_user?.birthDate),
      ),
      _div(),
      _row(Icons.favorite_outline, 'Agama', _user?.religion ?? '-'),
      _div(),
      _row(Icons.school_outlined, 'Pendidikan Terakhir', _user?.lastEducation ?? '-'),
      _div(),
      _row(Icons.calendar_month_outlined, 'Tahun Masuk', _user?.yearEnrolled ?? '-'),
      _div(),
      _row(Icons.work_outline, 'Status Kerja', _user?.employmentStatus ?? '-'),
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
    ]);
  }

  Widget _buildEmploymentInfo() {
    return _card('Detail Kepegawaian', Icons.work_outline, [
      _row(Icons.badge_outlined, 'NIK / Payroll', _user?.nik ?? '-'),
      _div(),
      _row(Icons.assignment_ind_outlined, 'Jabatan', _user?.position ?? '-'),
      _div(),
      _row(Icons.business_outlined, 'Departemen', _user?.department ?? '-'),
      _div(),
      _row(Icons.verified_user_outlined, 'Role', _user?.role ?? '-'),
      _div(),
      _row(
        Icons.toggle_on_outlined,
        'Status',
        (_user?.isActive ?? false) ? 'Aktif' : 'Non-Aktif',
      ),
    ]);
  }

  Widget _buildSettings() {
    return _card('Pengaturan', Icons.settings_outlined, [
      _settingTile(
        Icons.lock_outlined,
        'Ganti Password',
        onTap: _showChangePasswordDialog,
      ),
    ]);
  }

  Widget _card(String title, IconData icon, List<Widget> children) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 18, color: const Color(0xFF135BEC)),
              const SizedBox(width: 8),
              Text(
                title,
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF0F172A),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          ...children,
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
      padding: const EdgeInsets.symmetric(vertical: 7),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 17, color: Colors.grey.shade400),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade500,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 2),
                if (editable && ctrl != null)
                  TextField(
                    controller: ctrl,
                    maxLines: multi ? 3 : 1,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      color: Color(0xFF0F172A),
                    ),
                    decoration: InputDecoration(
                      isDense: true,
                      border: InputBorder.none,
                      hintText: 'Masukkan $label',
                      hintStyle: TextStyle(
                        color: Colors.grey.shade400,
                        fontSize: 13,
                      ),
                    ),
                  )
                else
                  Text(
                    value,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
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

  Widget _div() => Padding(
    padding: const EdgeInsets.symmetric(vertical: 2),
    child: Divider(color: Colors.grey.shade100, height: 1),
  );

  Widget _settingTile(
    IconData icon,
    String title, {
    Widget? trailing,
    VoidCallback? onTap,
  }) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(6),
        decoration: BoxDecoration(
          color: const Color(0xFFF1F5F9),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, size: 16, color: Colors.grey.shade600),
      ),
      title: Text(
        title,
        style: const TextStyle(fontSize: 13, color: Color(0xFF0F172A)),
      ),
      trailing:
          trailing ??
          const Icon(Icons.chevron_right, color: Colors.grey, size: 18),
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 0, vertical: 2),
      dense: true,
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
