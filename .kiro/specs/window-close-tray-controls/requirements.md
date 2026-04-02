# Requirements Document

## Introduction

Fitur ini menambahkan alur close app yang lebih aman dan jelas di CLIro-Go. Saat user menutup window utama, aplikasi harus menampilkan modal konfirmasi dengan pilihan untuk menutup aplikasi sepenuhnya atau meminimalkan aplikasi ke tray. Pada build Windows, tray juga harus menyediakan aksi cepat untuk membuka kembali window, mengaktifkan atau menonaktifkan API Router proxy, dan keluar dari aplikasi. Tujuannya adalah menjaga workflow desktop tetap aman, jelas, dan konsisten tanpa mematikan proxy secara tidak sengaja.

## Requirements

### Requirement 1

**User Story:** Sebagai user desktop, saya ingin aplikasi mengonfirmasi aksi close dari window utama, sehingga saya tidak menutup aplikasi atau proxy secara tidak sengaja.

#### Acceptance Criteria

1.1 WHEN user menekan tombol close window utama THEN sistem SHALL mencegat aksi close native sebelum aplikasi benar-benar keluar.
1.2 WHEN aksi close native berhasil dicegat THEN sistem SHALL menampilkan modal konfirmasi close yang konsisten dengan pola modal aplikasi yang sudah ada.
1.3 WHEN modal konfirmasi close sedang terbuka THEN sistem SHALL menjaga window utama tetap aktif dan aplikasi SHALL tetap berjalan.
1.4 IF user membatalkan atau menutup modal konfirmasi close tanpa memilih aksi akhir THEN sistem SHALL membatalkan proses close dan SHALL mempertahankan state aplikasi saat ini.

### Requirement 2

**User Story:** Sebagai user desktop, saya ingin memilih antara keluar penuh atau minimize ke tray, sehingga perilaku close sesuai dengan kebutuhan saya saat itu.

#### Acceptance Criteria

2.1 WHEN modal konfirmasi close ditampilkan THEN sistem SHALL menyediakan aksi `Close App` dan aksi `Minimize to Tray`.
2.2 WHEN user memilih `Close App` THEN sistem SHALL melakukan shutdown aplikasi penuh dan SHALL menjalankan alur cleanup backend yang sama seperti quit normal.
2.3 WHEN user memilih `Minimize to Tray` THEN sistem SHALL menyembunyikan window utama tanpa menghentikan runtime aplikasi, proxy, cloudflared, atau sesi auth yang masih aktif.
2.4 WHILE aplikasi sedang disembunyikan ke tray sistem SHALL tetap menjaga service backend yang sedang berjalan tetap tersedia.

### Requirement 3

**User Story:** Sebagai user Windows, saya ingin tray icon dan menu konteks saat aplikasi berjalan di background, sehingga saya tetap bisa mengontrol aplikasi tanpa membuka window utama terlebih dahulu.

#### Acceptance Criteria

3.1 WHERE build Windows sistem SHALL menyediakan tray icon selama runtime aplikasi aktif.
3.2 WHEN user membuka menu tray THEN sistem SHALL menampilkan aksi `Open` atau `Bring to Front`.
3.3 WHEN user membuka menu tray THEN sistem SHALL menampilkan aksi untuk mengaktifkan atau menonaktifkan API Router proxy sesuai state proxy saat ini.
3.4 WHEN user membuka menu tray THEN sistem SHALL menampilkan aksi `Exit App`.
3.5 WHEN tray icon atau aksi `Open` dipakai THEN sistem SHALL menampilkan kembali window utama ke depan layar dalam state yang dapat dipakai.
3.6 IF inisialisasi tray gagal pada build Windows THEN sistem SHALL mencatat kegagalan tersebut ke system log dan SHALL tetap menjaga aplikasi tetap dapat dipakai tanpa tray.

### Requirement 4

**User Story:** Sebagai user yang mengandalkan proxy lokal, saya ingin tray menu bisa mengontrol API Router proxy dengan aman, sehingga saya bisa mengelola proxy tanpa membuka tab API Router.

#### Acceptance Criteria

4.1 WHEN user memilih aksi tray untuk mengaktifkan proxy THEN sistem SHALL menjalankan alur start proxy yang sama dengan kontrol proxy utama aplikasi.
4.2 WHEN user memilih aksi tray untuk menonaktifkan proxy THEN sistem SHALL menjalankan alur stop proxy yang sama dengan kontrol proxy utama aplikasi.
4.3 WHEN state proxy berubah melalui tray THEN sistem SHALL memperbarui label atau status menu tray agar mencerminkan state proxy terbaru.
4.4 IF start atau stop proxy dari tray gagal THEN sistem SHALL mempertahankan state proxy sebelumnya dan SHALL mencatat error yang relevan.

### Requirement 5

**User Story:** Sebagai user yang kembali membuka aplikasi setelah aksi tray atau close interception, saya ingin UI tetap sinkron dengan state backend, sehingga saya langsung melihat status aplikasi yang akurat.

#### Acceptance Criteria

5.1 WHEN aplikasi dipulihkan dari tray THEN sistem SHALL memberi sinyal ke frontend untuk menyegarkan state inti aplikasi.
5.2 WHEN state proxy berubah melalui tray THEN sistem SHALL memberi sinyal ke frontend untuk menyegarkan `GetState()` dan `GetProxyStatus()`.
5.3 WHEN aksi `Exit App` dipilih dari tray THEN sistem SHALL keluar tanpa menampilkan ulang modal close konfirmasi.
5.4 WHEN window utama dipulihkan dari tray THEN sistem SHALL mempertahankan tab aktif, log subscription, dan session UI yang sedang berjalan sejauh runtime aplikasi belum di-shutdown.
