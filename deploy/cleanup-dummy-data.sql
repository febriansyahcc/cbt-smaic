-- ============================================================
-- CBT SMAS Islamic Centre Demak
-- Script: Bersihkan data dummy / percobaan dari semua master
-- Jalankan di server: docker exec -i cbt-postgres psql -U cbt_user -d cbt_db < cleanup-dummy-data.sql
-- PERINGATAN: Script ini PERMANEN. Backup dulu jika perlu.
-- ============================================================

BEGIN;

-- -----------------------------------------------------------
-- 1. Hapus akun guru/staf percobaan
--    Ganti username di bawah sesuai akun dummy di server Anda.
--    Akun 'admin' TIDAK dihapus (admin utama sistem).
-- -----------------------------------------------------------
DO $$
DECLARE
  dummy_usernames TEXT[] := ARRAY['guru1', 'guru2', 'guru_test', 'test'];
  uname TEXT;
  uid UUID;
  admin_id UUID;
  bank_count INT;
BEGIN
  -- Cari admin utama untuk menerima pemindahan bank soal
  SELECT id INTO admin_id FROM users WHERE username = 'admin' AND role = 'ADMIN' LIMIT 1;

  FOREACH uname IN ARRAY dummy_usernames LOOP
    SELECT id INTO uid FROM users WHERE username = uname AND role IN ('GURU', 'ADMIN');
    IF uid IS NOT NULL THEN
      -- Pindahkan bank soal ke admin agar FK tidak error
      SELECT COUNT(*) INTO bank_count FROM question_banks WHERE created_by_id = uid;
      IF bank_count > 0 THEN
        UPDATE question_banks SET created_by_id = admin_id WHERE created_by_id = uid;
        RAISE NOTICE 'Dipindahkan: % bank soal dari % ke admin', bank_count, uname;
      END IF;
      -- Hapus alokasi kelas mapel yang dipegang guru ini
      DELETE FROM class_subjects WHERE teacher_id = uid;
      -- Baru hapus user
      DELETE FROM users WHERE id = uid;
      RAISE NOTICE 'Dihapus: akun guru/staf dengan username %', uname;
    END IF;
  END LOOP;
END $$;

-- -----------------------------------------------------------
-- 2. Hapus kelas percobaan (yang tidak punya siswa dan tidak
--    terhubung jadwal ujian)
--    Sesuaikan daftar nama kelas dummy di bawah.
-- -----------------------------------------------------------
DO $$
DECLARE
  dummy_class_names TEXT[] := ARRAY['Test Class', 'Kelas Test', 'Demo'];
  cname TEXT;
  cid UUID;
  student_count INT;
  schedule_count INT;
BEGIN
  FOREACH cname IN ARRAY dummy_class_names LOOP
    SELECT id INTO cid FROM class_rooms WHERE name = cname;
    IF cid IS NOT NULL THEN
      SELECT COUNT(*) INTO student_count FROM student_profiles WHERE class_room_id = cid;
      SELECT COUNT(*) INTO schedule_count FROM exam_schedules WHERE class_room_id = cid;
      IF student_count = 0 AND schedule_count = 0 THEN
        DELETE FROM class_subjects WHERE class_room_id = cid;
        DELETE FROM class_rooms WHERE id = cid;
        RAISE NOTICE 'Dihapus: kelas %', cname;
      ELSE
        RAISE NOTICE 'Dilewati: kelas % (masih punya % siswa / % jadwal)', cname, student_count, schedule_count;
      END IF;
    END IF;
  END LOOP;
END $$;

-- -----------------------------------------------------------
-- 3. Hapus mata pelajaran percobaan (yang tidak punya bank soal
--    dan tidak terhubung kelas mapel)
-- -----------------------------------------------------------
DO $$
DECLARE
  dummy_subject_codes TEXT[] := ARRAY['TEST', 'DEMO', 'MAT-WJB-XII', 'BIN-WJB-XII'];
  scode TEXT;
  sid UUID;
  bank_count INT;
  alloc_count INT;
BEGIN
  FOREACH scode IN ARRAY dummy_subject_codes LOOP
    SELECT id INTO sid FROM subjects WHERE code = scode;
    IF sid IS NOT NULL THEN
      SELECT COUNT(*) INTO bank_count FROM question_banks WHERE subject_id = sid;
      SELECT COUNT(*) INTO alloc_count FROM class_subjects WHERE subject_id = sid;
      IF bank_count = 0 AND alloc_count = 0 THEN
        DELETE FROM subjects WHERE id = sid;
        RAISE NOTICE 'Dihapus: mata pelajaran kode %', scode;
      ELSE
        RAISE NOTICE 'Dilewati: mapel % (masih punya % bank soal / % alokasi)', scode, bank_count, alloc_count;
      END IF;
    END IF;
  END LOOP;
END $$;

-- -----------------------------------------------------------
-- 4. Hapus event ujian percobaan (beserta jadwal dan sesi terkait)
--    Ganti kode event sesuai yang ada di server Anda.
-- -----------------------------------------------------------
DO $$
DECLARE
  dummy_event_codes TEXT[] := ARRAY['CBT2026', 'ASAT26', 'TEST2026', 'DEMO2026'];
  ecode TEXT;
  eid UUID;
  sid UUID;
BEGIN
  FOREACH ecode IN ARRAY dummy_event_codes LOOP
    SELECT id INTO eid FROM exam_events WHERE code = ecode;
    IF eid IS NOT NULL THEN
      -- Hapus semua data dalam event (urutan FK-safe)
      FOR sid IN SELECT id FROM exam_schedules WHERE event_id = eid LOOP
        DELETE FROM violation_logs WHERE session_id IN (SELECT id FROM exam_sessions WHERE schedule_id = sid);
        DELETE FROM student_answers WHERE session_id IN (SELECT id FROM exam_sessions WHERE schedule_id = sid);
        DELETE FROM exam_sessions WHERE schedule_id = sid;
      END LOOP;
      DELETE FROM exam_schedules WHERE event_id = eid;
      DELETE FROM exam_events WHERE id = eid;
      RAISE NOTICE 'Dihapus: event ujian kode %', ecode;
    END IF;
  END LOOP;
END $$;

-- -----------------------------------------------------------
-- 5. Hapus alokasi kelas-mapel tanpa guru valid
-- -----------------------------------------------------------
DELETE FROM class_subjects
WHERE teacher_id NOT IN (SELECT id FROM users WHERE role IN ('GURU', 'ADMIN'));

-- -----------------------------------------------------------
-- 6. Ringkasan data yang tersisa
-- -----------------------------------------------------------
SELECT 'RINGKASAN SETELAH CLEANUP:' AS info;
SELECT 'Siswa aktif' AS entitas, COUNT(*)::TEXT AS jumlah FROM users WHERE role = 'SISWA' AND is_active = TRUE
UNION ALL
SELECT 'Guru & Staf', COUNT(*)::TEXT FROM users WHERE role IN ('GURU', 'ADMIN')
UNION ALL
SELECT 'Kelas (Rombel)', COUNT(*)::TEXT FROM class_rooms
UNION ALL
SELECT 'Mata Pelajaran', COUNT(*)::TEXT FROM subjects
UNION ALL
SELECT 'Alokasi Kelas-Mapel', COUNT(*)::TEXT FROM class_subjects
UNION ALL
SELECT 'Event Ujian', COUNT(*)::TEXT FROM exam_events
UNION ALL
SELECT 'Jadwal Ujian', COUNT(*)::TEXT FROM exam_schedules;

COMMIT;
