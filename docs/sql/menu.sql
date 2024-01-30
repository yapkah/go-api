-- phpMyAdmin SQL Dump
-- version 4.5.4.1deb2ubuntu2.1
-- http://www.phpmyadmin.net
--
-- Host: localhost
-- Generation Time: May 29, 2020 at 02:31 PM
-- Server version: 5.7.27-0ubuntu0.16.04.1
-- PHP Version: 7.0.33-0ubuntu0.16.04.6

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `db_gm_pac_vested_distribution`
--

-- --------------------------------------------------------

--
-- Table structure for table `sys_menu`
--

CREATE TABLE `slot_menu` (
  `id` int(11) NOT NULL,
  `name` varchar(50) DEFAULT NULL,
  `module_name` varchar(30) NOT NULL,
  `parent_id` int(11) NOT NULL,
  `level` int(11) NOT NULL,
  `seq_no` int(11) DEFAULT NULL,
  `file_path` varchar(100) NOT NULL,
  `status` varchar(30) NOT NULL,
  `created_by` varchar(30) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` varchar(30) DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
--
-- Indexes for dumped tables
--

--
-- Indexes for table `sys_menu`
--
ALTER TABLE `slot_menu`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `sys_menu`
--
ALTER TABLE `slot_menu`
    MODIFY `id` int NOT NULL AUTO_INCREMENT;
COMMIT;