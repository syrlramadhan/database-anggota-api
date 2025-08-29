-- MySQL dump 10.13  Distrib 8.0.43, for Linux (x86_64)
--
-- Host: localhost    Database: coconut_member
-- ------------------------------------------------------
-- Server version	8.0.43-0ubuntu0.22.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `angkatan`
--

DROP TABLE IF EXISTS `angkatan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `angkatan` (
  `id_angkatan` varchar(20) NOT NULL,
  `nama_angkatan` varchar(100) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id_angkatan`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `angkatan`
--

LOCK TABLES `angkatan` WRITE;
/*!40000 ALTER TABLE `angkatan` DISABLE KEYS */;
INSERT INTO `angkatan` VALUES ('000','Badan Pendiri','2025-08-07 19:44:43','2025-08-07 19:44:43'),('001','Angkatan 01','2025-08-07 19:44:43','2025-08-07 19:44:43'),('002','Angkatan 02','2025-08-07 19:44:43','2025-08-07 19:44:43'),('003','Angkatan 03','2025-08-07 19:44:43','2025-08-07 19:44:43'),('004','Angkatan 04','2025-08-07 19:44:43','2025-08-07 19:44:43'),('005','Angkatan 05','2025-08-07 19:44:43','2025-08-07 19:44:43'),('006','Angkatan 06','2025-08-07 19:44:43','2025-08-07 19:44:43'),('007','Angkatan 07','2025-08-07 19:44:43','2025-08-07 19:44:43'),('008','Angkatan 08','2025-08-07 19:44:43','2025-08-07 19:44:43'),('009','Angkatan 09','2025-08-07 19:44:43','2025-08-07 19:44:43'),('010','Angkatan 010','2025-08-07 19:44:43','2025-08-07 19:44:43'),('011','Angkatan 011','2025-08-07 19:44:43','2025-08-07 19:44:43'),('012','Angkatan 012','2025-08-07 19:44:43','2025-08-07 19:44:43'),('013','Angkatan 013','2025-08-07 19:44:43','2025-08-07 19:44:43'),('014','Angkatan 014','2025-08-07 19:44:43','2025-08-07 19:44:43');
/*!40000 ALTER TABLE `angkatan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `jurusan`
--

DROP TABLE IF EXISTS `jurusan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `jurusan` (
  `id_jurusan` varchar(10) NOT NULL,
  `nama_jurusan` varchar(100) NOT NULL,
  PRIMARY KEY (`id_jurusan`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `jurusan`
--

LOCK TABLES `jurusan` WRITE;
/*!40000 ALTER TABLE `jurusan` DISABLE KEYS */;
INSERT INTO `jurusan` VALUES ('J001','Frontend'),('J002','Backend'),('J003','System');
/*!40000 ALTER TABLE `jurusan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `member`
--

DROP TABLE IF EXISTS `member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `member` (
  `id_member` varchar(36) NOT NULL,
  `nra` varchar(10) DEFAULT NULL,
  `nama` varchar(100) NOT NULL,
  `angkatan` varchar(20) NOT NULL,
  `status_keanggotaan` enum('anggota','bph','alb','dpo','bp') NOT NULL,
  `id_jurusan` varchar(10) DEFAULT NULL,
  `tanggal_dikukuhkan` varchar(15) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `no_hp` varchar(20) DEFAULT NULL,
  `password` varchar(100) DEFAULT NULL,
  `foto` varchar(100) DEFAULT NULL,
  `login_token` varchar(100) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id_member`),
  UNIQUE KEY `nra` (`nra`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `login_token` (`login_token`),
  KEY `angkatan` (`angkatan`),
  KEY `id_jurusan` (`id_jurusan`),
  CONSTRAINT `member_ibfk_1` FOREIGN KEY (`angkatan`) REFERENCES `angkatan` (`id_angkatan`),
  CONSTRAINT `member_ibfk_2` FOREIGN KEY (`id_jurusan`) REFERENCES `jurusan` (`id_jurusan`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `member`
--

LOCK TABLES `member` WRITE;
/*!40000 ALTER TABLE `member` DISABLE KEYS */;
INSERT INTO `member` VALUES ('0efe56f3-2142-49ef-ba19-4739adbba139','12.23.001','ABD RAHMAN WAHID','012','dpo','J003','2023-06-13','maman@gmail.com','085341847801','$2a$10$A2gXAfG5KrptpnlmTM8YoO/mR/BoKCa3bNY4WfuaW6/QLSDj/s2yy','12.23.001_ABD RAHMAN WAHID.jpg',NULL,'2025-08-21 18:32:52','2025-08-21 18:32:52'),('56a8090d-36da-4e54-86f0-1599c37d0c49','13.24.005','Syahrul Ramadhan','013','bph','J002','2025-08-28','syahrul@gmail.com','085371847801','$2a$10$mJaNda8cNP8CdkXyHXze3e2SYSoP/MarHzLvYA/9Y1c6gzB8ACdQ6','13.24.005_Syahrul Ramadhan.jpg',NULL,'2025-08-28 18:31:57','2025-08-28 18:31:57');
/*!40000 ALTER TABLE `member` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-08-28 19:58:07
