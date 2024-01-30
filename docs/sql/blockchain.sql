-- phpMyAdmin SQL Dump
-- version 4.6.6deb5
-- https://www.phpmyadmin.net/
--
-- Host: 10.1.3.151:3306
-- Generation Time: Jun 10, 2020 at 01:58 PM
-- Server version: 5.7.30-0ubuntu0.18.04.1-log
-- PHP Version: 7.2.24-0ubuntu0.18.04.6

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `db_gm_live`
--

-- --------------------------------------------------------

--
-- Table structure for table `blockchain_api_log`
--

CREATE TABLE `blockchain_api_log` (
  `id` int(11) NOT NULL,
  `prj_config_code` varchar(25) DEFAULT NULL,
  `side` varchar(25) NOT NULL,
  `api_type` varchar(50) NOT NULL,
  `method` varchar(25) DEFAULT NULL,
  `url_link` varchar(255) NOT NULL,
  `data_sent` text,
  `data_received` text,
  `server_data` varchar(255) DEFAULT NULL,
  `dt_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `blockchain_api_log`
--
ALTER TABLE `blockchain_api_log`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `blockchain_api_log`
--
ALTER TABLE `blockchain_api_log`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;