-- phpMyAdmin SQL Dump
-- version 4.8.3
-- https://www.phpmyadmin.net/
--
-- Host: localhost
-- Generation Time: May 18, 2020 at 04:46 PM
-- Server version: 10.1.35-MariaDB
-- PHP Version: 7.2.9

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `gta-api`
--

-- --------------------------------------------------------

--
-- Table structure for table `ewt_deposit`
--

CREATE TABLE `ewt_deposit` (
    `id` int(11) NOT NULL,
    `member_id` int(11) NOT NULL,
    `ewallet_type_id` int(11) NOT NULL,
    `trans_date` varchar(50) DEFAULT NULL,
    `doc_no` varchar(100) NOT NULL,
    `amount` decimal(25,10) DEFAULT '0.0000000000',
    `status` varchar(5) DEFAULT 'A',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `ewt_withdrawal`
--

CREATE TABLE `ewt_withdrawal` (
    `id` int(11) NOT NULL,
    `member_id` int(11) NOT NULL,
    `ewallet_type_id` int(11) NOT NULL,
    `trans_date` varchar(50) DEFAULT NULL,
    `doc_no` varchar(100) NOT NULL,
    `amount` decimal(25,10) DEFAULT '0.0000000000',
    `status` varchar(5) DEFAULT 'A',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `ewt_transfer`
--

CREATE TABLE `ewt_transfer` (
    `id` int(11) NOT NULL,
    `member_from` int(11) NOT NULL,
    `member_to` int(11) NOT NULL,
    `ewallet_type_id` int(11) NOT NULL,
    `trans_date` varchar(50) DEFAULT NULL,
    `doc_no` varchar(100) NOT NULL,
    `amount` decimal(25,10) DEFAULT '0.0000000000',
    `status` varchar(5) DEFAULT 'A',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `ewt_detail`
--

CREATE TABLE `ewt_detail` (
    `id` int(11) NOT NULL,
    `member_id` int(11) NOT NULL,
    `ewallet_type_id` int(11) NOT NULL,
    `trans_date` varchar(50) DEFAULT NULL,
    `trans_type` varchar(50) DEFAULT NULL,
    `total_in` decimal(25,10) DEFAULT '0.0000000000',
    `total_out` decimal(25,10) DEFAULT '0.0000000000',
    `balance` decimal(25,10) DEFAULT '0.0000000000',
    `additional_remark` varchar(255) DEFAULT NULL,
    `remark` varchar(255) DEFAULT NULL,
    `ewt_deposit_id` int(11) NOT NULL,
    `ewt_withdrawal_id` int(11) NOT NULL,
    `ewt_transfer_id` int(11) NOT NULL,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `ewt_summary`
--

CREATE TABLE `ewt_summary` (
    `id` int(11) NOT NULL,
    `member_id` int(11) NOT NULL,
    `ewallet_type_id` int(11) NOT NULL,
    `total_in` decimal(25,10) DEFAULT '0.0000000000',
    `total_out` decimal(25,10) DEFAULT '0.0000000000',
    `balance` decimal(25,10) DEFAULT '0.0000000000',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `sys_doc_no`
--

CREATE TABLE `sys_doc_no` (
  `id` int(11) NOT NULL,
  `module` varchar(50) NOT NULL,
  `doc_type` varchar(50) NOT NULL,
  `doc_no_prefix` varchar(10) NOT NULL,
  `start_no` int(11) NOT NULL,
  `doc_length` int(11) DEFAULT '0',
  `running_no` varchar(100) NOT NULL,
  `running_type` varchar(100) NOT NULL,
  `table_name` varchar(100) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `ewt_setup`
--

CREATE TABLE `ewt_setup` (
  `id` int(11) NOT NULL,
  `ewallet_type_id` int(11) NOT NULL,
  `currency_code` varchar(10) NOT NULL,
  `b_deposit` tinyint(4) DEFAULT '0',
  `b_withdrawal` tinyint(4) DEFAULT '0',
  `b_transfer` tinyint(4) DEFAULT '0',
  `decimal_point` int(11) DEFAULT '8',
  `status` varchar(10) DEFAULT 'A',
  `withdraw_min` decimal(25,10) DEFAULT '0.0000000000',
  `withdraw_max` decimal(25,10) DEFAULT '0.0000000000',
  `transfer_min` decimal(25,10) DEFAULT '0.0000000000',
  `transfer_max` decimal(25,10) DEFAULT '0.0000000000',
  `seq_no` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `sys_general`
--

CREATE TABLE `sys_general` (
    `id` int(11) NOT NULL,
    `type` varchar(100) NOT NULL,
    `code` varchar(100) NOT NULL,
    `name` varchar(100) NOT NULL,
    `b_display_code` varchar(100) NOT NULL,
    `status` varchar(10) DEFAULT 'A',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `ewt_deposit`
--
ALTER TABLE `ewt_deposit`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `ewt_withdrawal`
--
ALTER TABLE `ewt_withdrawal`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `ewt_transfer`
--
ALTER TABLE `ewt_transfer`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `ewt_detail`
--
ALTER TABLE `ewt_detail`
  ADD PRIMARY KEY (`id`);


--
-- Indexes for table `ewt_summary`
--
ALTER TABLE `ewt_summary`
    ADD PRIMARY KEY (`id`);

--
-- Indexes for table `ewt_setup`
--
ALTER TABLE `ewt_setup`
    ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sys_doc_no`
--
ALTER TABLE `sys_doc_no`
    ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sys_general`
--
ALTER TABLE `sys_general`
    ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `ewt_deposit`
--
ALTER TABLE `ewt_deposit`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `ewt_withdrawal`
--
ALTER TABLE `ewt_withdrawal`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `ewt_transfer`
--
ALTER TABLE `ewt_transfer`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `ewt_summary`
--
ALTER TABLE `ewt_summary`
    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `ewt_detail`
--
ALTER TABLE `ewt_detail`
    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `ewt_setup`
--
ALTER TABLE `ewt_setup`
    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `sys_general`
--
ALTER TABLE `sys_general`
    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `sys_doc_no`
--
ALTER TABLE `sys_doc_no`
    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
