-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- 主机： mysql-a.fanhaninfo.test
-- 生成日期： 2023-08-01 03:17:38
-- 服务器版本： 8.0.29
-- PHP 版本： 8.1.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `dev_account_center`
--

-- --------------------------------------------------------

--
-- 表的结构 `accounts`
--

CREATE TABLE `accounts` (
  `id` int NOT NULL,
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `salt` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `can_delete` tinyint(1) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `accounts`
--

INSERT INTO `accounts` (`id`, `username`, `nickname`, `password`, `email`, `salt`, `can_delete`, `created_at`, `updated_at`) VALUES
(1, 'superadmin', '超级管理员', 'd445a3c28088b38d9d7d037201255109', 'aa', 'sUOHoF', 0, 0, 1667878497),
(2, 'admin', '管理员', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 0, 0, 1681106639),
(158, 'admin1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667878550),
(160, 'admin3', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(161, 'admin14', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(162, 'admin51', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(163, 'admin61', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(164, 'admin71', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(165, 'admin81', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(166, 'admin91', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(167, 'admin111', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(169, 'ad13min1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(170, 'adm31in1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(171, 'admi123n1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(172, 'ad555min1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(173, 'admian1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(174, 'admsin1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(175, 'admidn1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(176, 'admifn1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(177, 'admihgn1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(178, 'adtmin1', 'ccccc', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 1, 0, 1667816324),
(179, 'superadmina', 'ccccc', 'e3c977f168452ee0dbea84a38700e160', 'aa', 'sRFta7', 1, 1667820985, 1667820985),
(180, 'superadmi2na', 'ccccc', '50d13968c13ccf6c2aa544752a13a418', 'aa', 'CvtnZ2', 1, 1667821098, 1667821098),
(181, 'test', 'Test', '30cbd33a8ddfe01e3063c1a070a0f36a', '', 'Ig8J2m', 1, 1667877196, 1676865296),
(182, 'test2', 'test2', '663d09a06d46fd3484ead08a5fc28601', '', 'T5yDRg', 1, 1667877395, 1676617537),
(184, 'superadmi2na2', 'ccccc', '631ef5de37bda2a7b11a806b7c6aa9d4', 'aa', 'a4KO4D', 1, 1667877719, 1667877719),
(185, 'superadmi2na2a', 'ccccc', 'b6dc8392381604a20135d4071e4c6fc1', 'aa', 'Yo4NhD', 1, 1667877834, 1667877834),
(186, 'test3', 'test32', 'c5823d84eb6e3ef9313db8fc31df7719', '123', 'uF33lT', 1, 1667877975, 1667878108),
(187, 'myadmin', '我的管理员', '200f78448ca677e03c94c121ae55a211', '', 'fvDMPb', 1, 1668131711, 1672392140),
(188, 'AAA', 'AAA', 'd77efad6eb712b0968a03e554beba922', '', 'fzJQOq', 1, 1668132126, 1668132126),
(189, 'lvyitao', 'lvyitao', '2c98ecb63ca0fdad56220f46d2e13995', '', 'UhiQa1', 1, 1670815568, 1688543633),
(190, 'guest', 'Guest', '9f55a5d090b2c7142d47e13a0e169d59', '', '0zBi07', 1, 1670983145, 1678332729),
(191, 'chenzhong', '陈中', 'd4c05de4de7b408133d06474834e347e', '', 'XYUMVY', 1, 1670983169, 1676970088),
(192, 'gaoting', '高婷', '9673e73ac84bd5c2c80b7591185a77b5', '', 'ftqe2z', 1, 1673852250, 1673852300),
(193, 'wangzengming', '王增明', 'c71bdad712fa4c94d9edb3ea71181e15', '', 'QFsjEh', 1, 1676008018, 1677640703),
(194, 'sales01', '销售01', '4327023fe1abdadd21c69399c5f12b72', '', 'GHV6Wh', 1, 1678861858, 1678862031),
(195, 'license_server', '许可证管理员', '9f9192bc73c1b4f62b4bea6b7170b65a', '', '5L7MPw', 1, 1680744833, 1680744833),
(198, 'testlv', 'testlv', 'a414b7e9728688d7c281fe7ba1939a85', '', 'rZmT5z', 1, 1688619262, 1688619262);

-- --------------------------------------------------------

--
-- 表的结构 `accounts_token`
--

CREATE TABLE `accounts_token` (
  `token_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `uid` int NOT NULL,
  `expiration` int NOT NULL,
  `updated_at` int NOT NULL,
  `created_at` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `accounts_token`
--

INSERT INTO `accounts_token` (`token_key`, `uid`, `expiration`, `updated_at`, `created_at`) VALUES
('15e0dc16afc3cddc00a6bbbb9e827a66', 2, 1690941546, 1689213546, 1689213546),
('5160edd4247ec56ff6ea00ad02da3c1d', 2, 1692080762, 1690352762, 1690352762),
('64084d36951e74206c522cddb4ad45e5', 191, 1691033891, 1689305891, 1689305891),
('851570d1fe6abf7ea02cbbd97cbf080c', 2, 1691670330, 1689942330, 1689942330),
('949200ab4db10adcfe19cf79f53da385', 2, 1691205197, 1689477197, 1689477197),
('b24172f7daa7216c18103fa4b43b9eb7', 191, 1691033428, 1689305428, 1689305428),
('ddbe50af43dd5b60aa9e9127369e1340', 191, 1691028934, 1689300934, 1689300934),
('e4a1e91ecd0142bff8ba7f395a3a69e0', 191, 1691031774, 1689303774, 1689303774),
('ed0b93edb62a2323bfe507fcf66f161f', 191, 1691033259, 1689305259, 1689305259),
('f6c37c79e5f9672d132c8e1f849bb97b', 191, 1691028718, 1689300718, 1689300718);

-- --------------------------------------------------------

--
-- 表的结构 `departments`
--

CREATE TABLE `departments` (
  `id` int NOT NULL COMMENT '部门id',
  `fid` bigint DEFAULT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '部门查找key',
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '部门名称（显示用）',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '备注',
  `updated_at` bigint NOT NULL COMMENT '更新时间',
  `created_at` bigint NOT NULL COMMENT '创建时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `departments`
--

INSERT INTO `departments` (`id`, `fid`, `name`, `nickname`, `remark`, `updated_at`, `created_at`) VALUES
(1, 0, 'zhaiquanyewuguanlizongbu', '债券业务管理总部', '', 1667266057, 0),
(162, 0, 'gudingshouyizongbu', '固定收益总部', '', 1667214752, 0),
(163, 162, 'xiansibu', '西安四部', '', 1666690385, 0),
(164, 0, 'chanyejinrongbu', '产业金融部', '', 1667530636, 0),
(165, 0, 'qiyerongzibu', '企业融资部', '', 1666690356, 0),
(166, 0, 'xinxijishuzhongxin', '信息技术中心', '', 1667530637, 0),
(167, 0, 'jinrongkejibu', '金融科技部', '', 1667266316, 0),
(168, 0, 'xinxijishuzhongxin', '信息技术中心', '', 1667380442, 0),
(1685, 0, 'kysec', '开源证券', '', 1666765810, 1666765810),
(1686, 162, 'xawb', '西安五部', '', 1666766047, 0),
(5264, 0, 'gudingshouyizongbu', '固定收益总部', '', 1667530635, 1667381043),
(5265, 5264, 'xiansibu', '西安四部', '', 1667530636, 1667381043),
(5267, 0, 'qiyerongzibu', '企业融资部', '', 1667530637, 1667381044),
(5268, 0, 'jinrongkejibu', '金融科技部', '', 1667530637, 1667381045);

-- --------------------------------------------------------

--
-- 表的结构 `departments_accounts`
--

CREATE TABLE `departments_accounts` (
  `id` int NOT NULL,
  `account_id` int NOT NULL,
  `department_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `departments_accounts`
--

INSERT INTO `departments_accounts` (`id`, `account_id`, `department_id`) VALUES
(9, 185, 1),
(12, 186, 1),
(14, 1, 1),
(15, 158, 1),
(22, 187, 162),
(23, 187, 163),
(39, 192, 1),
(40, 192, 162),
(41, 192, 163),
(42, 192, 164),
(43, 192, 165),
(44, 192, 166),
(50, 193, 164),
(56, 2, 1);

-- --------------------------------------------------------

--
-- 表的结构 `departments_jurisdictions`
--

CREATE TABLE `departments_jurisdictions` (
  `id` int NOT NULL COMMENT '权限流水id',
  `department_id` int NOT NULL COMMENT '所属部门',
  `jurisdiction_id` int NOT NULL COMMENT '权限id（索引）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `identitys`
--

CREATE TABLE `identitys` (
  `id` int NOT NULL,
  `code` varchar(255) NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `is_system` int NOT NULL DEFAULT '0',
  `updated_at` bigint NOT NULL,
  `created_at` bigint NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- 转存表中的数据 `identitys`
--

INSERT INTO `identitys` (`id`, `code`, `name`, `remark`, `is_system`, `updated_at`, `created_at`) VALUES
(1, 'superadmin', '超级管理员', '', 1, 0, 0),
(25, 'department-manager', '部门管理员', '', 0, 1670983632, 1670983632),
(26, 'role-manager', '角色管理员', '', 0, 1670983655, 1670983655),
(27, 'user-manager', '用户管理员', '', 0, 1670983672, 1670983672),
(31, 'aaa', 'aaa', '', 0, 1675761713, 1675761713),
(32, 'ai', 'ai_server', '', 0, 1675995073, 1675995073),
(33, 'digital_employee_manager', '数字化员工管理员', '数字化员工管理员2', 0, 1677640715, 1676007987),
(34, 'default_role', '默认角色（请勿删除）', '', 0, 1677636283, 1676013736),
(35, 'viewer', '查看者', '', 0, 1676343198, 1676343198),
(36, 'hyperautomation_manager', '超自动化平台管理员', '', 0, 1676451031, 1676449274),
(37, 'manage', '管理', '', 0, 1677571260, 1677571260),
(38, 'business', '业务', '', 0, 1677634637, 1677634637),
(39, 'sales', '销售', '销售', 0, 1681695673, 1678861878),
(40, 'license_adminstrator', '许可证管理员', '', 0, 1681106902, 1680744774),
(41, 'ope_enginerr', '许可证模块运维工程师', '', 0, 1681462660, 1681462660),
(42, 'test', 'test', '', 0, 1688720203, 1688611707);

-- --------------------------------------------------------

--
-- 表的结构 `identitys_accounts`
--

CREATE TABLE `identitys_accounts` (
  `id` int NOT NULL,
  `identity_id` int NOT NULL,
  `account_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `identitys_accounts`
--

INSERT INTO `identitys_accounts` (`id`, `identity_id`, `account_id`) VALUES
(42, 1, 185),
(43, 2, 185),
(46, 1, 186),
(47, 2, 186),
(48, 3, 186),
(52, 1, 1),
(53, 2, 1),
(54, 1, 158),
(88, 1, 192),
(89, 25, 192),
(90, 26, 192),
(91, 27, 192),
(143, 1, 191),
(144, 25, 191),
(145, 26, 191),
(146, 27, 191),
(148, 32, 191),
(149, 36, 191),
(154, 36, 182),
(160, 36, 181),
(195, 33, 193),
(196, 34, 193),
(197, 36, 193),
(198, 37, 193),
(199, 39, 194),
(200, 34, 195),
(201, 40, 195),
(221, 1, 2),
(222, 40, 2),
(234, 25, 189),
(235, 26, 189),
(236, 27, 189),
(237, 32, 189),
(238, 42, 198);

-- --------------------------------------------------------

--
-- 表的结构 `identitys_jurisdictions`
--

CREATE TABLE `identitys_jurisdictions` (
  `id` int NOT NULL,
  `identity_id` int NOT NULL COMMENT '所属身份',
  `jurisdiction_id` int NOT NULL COMMENT '权限id（索引）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- 转存表中的数据 `identitys_jurisdictions`
--

INSERT INTO `identitys_jurisdictions` (`id`, `identity_id`, `jurisdiction_id`) VALUES
(1, 1, 1),
(11, 2, 1),
(12, 1, 2),
(50, 25, 11),
(51, 25, 12),
(52, 25, 13),
(53, 25, 16),
(54, 26, 6),
(55, 26, 7),
(56, 26, 8),
(57, 26, 15),
(58, 27, 2),
(59, 27, 3),
(60, 27, 4),
(61, 27, 14),
(65, 1, -1),
(67, 31, 34),
(68, 31, 35),
(69, 31, 36),
(70, 32, 18),
(71, 32, 17),
(247, 35, 95),
(268, 36, 93),
(269, 36, 94),
(270, 36, 97),
(271, 36, 106),
(272, 36, 95),
(273, 36, 96),
(274, 36, 98),
(275, 36, 100),
(276, 36, 101),
(277, 36, 102),
(278, 36, 103),
(279, 36, 104),
(280, 36, 105),
(281, 36, 107),
(282, 36, 108),
(283, 36, 109),
(284, 36, 110),
(285, 36, 111),
(286, 36, 116),
(287, 36, 117),
(288, 36, 112),
(289, 36, 113),
(290, 36, 114),
(291, 36, 115),
(320, 37, 125),
(321, 38, 124),
(372, 34, 90),
(373, 34, 91),
(374, 34, 92),
(375, 34, 118),
(376, 34, 119),
(377, 34, 88),
(598, 33, 45),
(599, 33, 47),
(600, 33, 52),
(601, 33, 63),
(602, 33, 77),
(603, 33, 82),
(604, 33, 127),
(605, 33, 35),
(606, 33, 40),
(607, 33, 48),
(608, 33, 49),
(609, 33, 50),
(610, 33, 51),
(611, 33, 53),
(612, 33, 54),
(613, 33, 55),
(614, 33, 56),
(615, 33, 57),
(616, 33, 58),
(617, 33, 59),
(618, 33, 60),
(619, 33, 61),
(620, 33, 62),
(621, 33, 64),
(622, 33, 65),
(623, 33, 66),
(624, 33, 78),
(625, 33, 79),
(626, 33, 80),
(627, 33, 81),
(628, 33, 83),
(629, 33, 84),
(630, 33, 85),
(631, 33, 87),
(632, 33, 128),
(633, 33, 129),
(634, 33, 130),
(635, 33, 131),
(636, 33, 37),
(637, 33, 38),
(638, 33, 39),
(639, 33, 41),
(640, 33, 42),
(641, 33, 43),
(642, 33, 44),
(643, 33, 67),
(644, 33, 68),
(645, 33, 71),
(646, 33, 72),
(647, 33, 73),
(648, 33, 74),
(649, 33, 75),
(650, 33, 76),
(651, 33, 69),
(652, 33, 70),
(732, 40, 20),
(733, 40, 21),
(734, 40, 22),
(735, 40, 137),
(736, 40, 144),
(737, 40, 138),
(738, 40, 141),
(739, 40, 145),
(740, 40, 140),
(741, 40, 142),
(742, 40, 146),
(743, 40, 143),
(744, 41, 136),
(756, 39, 20),
(757, 39, 21),
(758, 39, 138),
(759, 39, 140),
(760, 39, 141),
(761, 39, 142),
(762, 39, 143),
(763, 39, 137),
(764, 39, 145),
(765, 39, 146),
(766, 39, 144),
(767, 39, 135),
(768, 39, 22),
(788, 42, 30),
(801, 42, 178),
(804, 42, 181),
(805, 42, 182);

-- --------------------------------------------------------

--
-- 表的结构 `jurisdictions`
--

CREATE TABLE `jurisdictions` (
  `id` int NOT NULL,
  `code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `fid` bigint DEFAULT '0',
  `is_system` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- 转存表中的数据 `jurisdictions`
--

INSERT INTO `jurisdictions` (`id`, `code`, `name`, `fid`, `is_system`) VALUES
(-1, 'superadmin', '超级管理员', 0, 1),
(1, 'account', '账户模块', 0, 0),
(2, 'account_create', '新增账户', 1, 0),
(3, 'account_delete', '账户删除', 1, 0),
(4, 'account_update', '账户修改', 1, 0),
(5, 'identity', '角色模块', 0, 0),
(6, 'identity_create', '新增角色', 5, 0),
(7, 'identity_delete', '删除角色', 5, 0),
(8, 'identity_update', '修改角色', 5, 0),
(10, 'department', '部门模块', 0, 0),
(11, 'department_create', '新增部门', 10, 0),
(12, 'department_delete', '删除部门', 10, 0),
(13, 'department_update', '修改部门', 10, 0),
(14, 'account_list', '账户列表', 1, 0),
(15, 'identity_list', '角色列表', 5, 0),
(16, 'department_list', '部门列表', 10, 0),
(17, 'ai_service', 'AI模块', 0, 0),
(18, 'ai_service_process', 'AI_Process', 17, 0),
(19, 'license', '许可证模块', 0, 0),
(20, 'license_order_create', '下单', 19, 0),
(21, 'license_order_list', '订单管理', 19, 0),
(22, 'license_list', '许可证管理', 19, 0),
(23, 'jsonrpc', '内部RPC', 0, 0),
(24, 'jsonrpc_call', '发起请求', 23, 0),
(26, 'digital_employee_process', '查看当前部门流程数据', 25, 0),
(27, 'digital_employee_process_list_data', '流程列表数据权限', 26, 0),
(30, 'test', '测试', 0, 0),
(34, 'digital_employee_platform', '数字员工平台', 0, 0),
(35, 'process', '定制流程', 45, 0),
(37, '/process/update', '维护', 35, 0),
(38, '/process/delete', '删除', 35, 0),
(39, '/process/detail', '详情', 35, 0),
(40, 'process_generate', '通用流程', 45, 0),
(41, '/process/list', '列表', 35, 0),
(42, '/process_generate/detail', '详情', 40, 0),
(43, '/process_generate/update', '维护', 40, 0),
(44, '/process_generate/list', '列表', 40, 0),
(45, 'process_lib', '流程库', 34, 0),
(47, 'digital_employee', '数字化员工', 34, 0),
(48, '/digital_employee/update', '维护', 47, 0),
(49, '/digital_employee/list', '列表', 47, 0),
(50, '/digital_employee/delete', '删除', 47, 0),
(51, '/digital_employee/detail', '详情', 47, 0),
(52, 'demand_manage', '需求管理', 34, 0),
(53, '/employeeClaims/needsApply', '发布需求', 52, 0),
(54, '/employeeClaims/dev', '开发者提交', 52, 0),
(55, '/employeeClaims/devOpsAdminEvaluate', '运维管理员审批', 52, 0),
(56, '/employeeClaims/coeAdminEvaluate', 'COE管理员审批', 52, 0),
(57, '/employeeClaims/devOpsAdminInspectionAcception', '运维管理员验收', 52, 0),
(58, '/employeeClaims/coeAdminInspectionAcception', 'COE管理员验收', 52, 0),
(59, '/employeeClaims/list', '需求列表', 52, 0),
(60, '/employeeClaims/detail', '需求详情', 52, 0),
(61, '/employeeClaims/recall', '需求撤回', 52, 0),
(62, '/employeeClaims/type/list', ' 需求类型列表', 52, 0),
(63, 'job_manage', '任务管理', 34, 0),
(64, 'job_search', '任务查询', 63, 0),
(65, 'immediate_job', '即时任务', 63, 0),
(66, 'schedule_job', '预约任务', 63, 0),
(67, '/job/list', '任务列表', 64, 0),
(68, '/job/detail', '查看详情', 64, 0),
(69, '/job_log/list', '任务日志列表', 68, 0),
(70, '/job/parameters', '任务参数列表', 68, 0),
(71, '/job/create', '创建任务', 65, 0),
(72, '/job/restart', '重试任务', 65, 0),
(73, '/job/stop', '停止任务', 65, 0),
(74, '/process_schedule/list', '任务列表', 66, 0),
(75, '/process_schedule/create', '创建任务', 66, 0),
(76, '/process_schedule/delete', '取消任务', 66, 0),
(77, 'dispatch_source_manage', '调度源管理', 34, 0),
(78, '/dispatch_source/list', '调度源列表', 77, 0),
(79, '/dispatch_source/create', '创建调度源', 77, 0),
(80, '/dispatch_source/update', '编辑调度源', 77, 0),
(81, '/dispatch_source/delete', '删除调度源', 77, 0),
(82, 'new_manage', '新闻管理', 34, 0),
(83, '/news/list', '新闻列表', 82, 0),
(84, '/news/create', '创建新闻', 82, 0),
(85, '/news/update', '修改新闻', 82, 0),
(87, '/news/detail', '新闻详情', 82, 0),
(88, 'interface_test', '接口测试', 0, 0),
(90, 'account_module', '用户模块', 88, 0),
(91, '/account/list', '用户列表', 90, 0),
(92, '/account/department/list', '部门列表', 90, 0),
(93, 'hyperautomation_platform', '超自动化平台', 0, 0),
(94, 'view_perm', '视图权限', 93, 1),
(95, 'view_inspect', '查看权限', 94, 1),
(96, 'view_manage', '管理权限', 94, 1),
(97, 'data_table', '业务数据表管理', 93, 0),
(98, '/data_table/create', '新增业务数据表', 97, 0),
(100, '/data_table/update', '编辑业务数据表', 97, 0),
(101, '/data_table/delete', '删除业务数据表', 97, 0),
(102, '/data_table/detail', '业务数据表详情', 97, 0),
(103, '/data_table/list', '业务数据表列表', 97, 0),
(104, '/data_table/add_field', '新增业务数据表字段', 97, 0),
(105, '/data_table/delete_field', '删除业务数据表字段', 97, 0),
(106, 'data_view_manage', '数据视图管理', 93, 0),
(107, '/data_view/list', '数据视图列表', 106, 0),
(108, '/data_view/create', '创建数据视图', 106, 0),
(109, '/data_view/update', '编辑数据视图', 106, 0),
(110, '/data_view/detail', '数据视图详情', 106, 0),
(111, 'permission_manage', '权限管理', 106, 0),
(112, '/data_view/permission/list', '权限列表', 111, 0),
(113, '/data_view/permission/add', '创建权限', 111, 0),
(114, '/data_view/permission/update', '编辑权限', 111, 0),
(115, '/data_view/permission/delete', '删除权限', 111, 0),
(116, '/data_view/data/list', '数据列表', 106, 0),
(117, '/data_view/get_secret', '获取密钥', 106, 0),
(118, '/account/identity/list', '角色列表', 90, 0),
(119, '/account/identity/list_by_jurisdiction_code', '权限编码获取角色列表', 90, 0),
(120, 'system_manage', '系统权限', 34, 1),
(121, 'admin', '管理员', 120, 1),
(122, 'coe', 'COE', 120, 1),
(123, 'dev', '开发', 120, 1),
(124, 'business', '业务', 120, 1),
(125, 'devops_manager', '运维管理', 120, 1),
(126, 'guest', '访客', 120, 1),
(127, 'label_manage', '标签管理', 34, 0),
(128, '/label/list', '标签列表', 127, 0),
(129, '/label/create', '创建标签', 127, 0),
(130, '/label/update', '编辑标签', 127, 0),
(131, '/label/delete', '删除标签', 127, 0),
(132, 'coe_manager', 'COE管理', 120, 0),
(133, 'system_prev', '系统权限', 19, 0),
(134, 'license_admin', '管理员', 133, 0),
(135, 'license_sales', '销售', 133, 0),
(136, 'license_operator', '运维工程师', 133, 0),
(137, 'product_manage', '产品管理', 19, 0),
(138, '/product/list', '产品列表', 137, 0),
(140, '/order/sale/create', '下单', 138, 0),
(141, 'order_manage', '订单管理', 137, 0),
(142, '/order/sale/list', '订单列表', 141, 0),
(143, '/order/sale/detail', '订单详情', 142, 0),
(144, 'license_manage', '许可证管理', 19, 0),
(145, '/license/list', '许可证列表', 144, 0),
(146, '/license/detail', '许可证详情', 145, 0),
(149, 'testauth01', 'testauth01', 148, 0),
(151, 'nimbleAI', 'AI_NimbleAI', 17, 0),
(156, 'test111', 'test111', 0, 0),
(178, 'test1_1', 'test1_1', 30, 0),
(181, 'test1_1_1', 'test1_1_1', 178, 0),
(182, 'test2_1', 'test2_1', 30, 0);

-- --------------------------------------------------------

--
-- 表的结构 `locker`
--

CREATE TABLE `locker` (
  `guid` varchar(255) DEFAULT NULL,
  `method` varchar(255) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `logs`
--

CREATE TABLE `logs` (
  `id` bigint NOT NULL,
  `type` varchar(255) DEFAULT NULL,
  `msg` varchar(255) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- 转存表中的数据 `logs`
--

INSERT INTO `logs` (`id`, `type`, `msg`, `created_at`, `updated_at`) VALUES
(2311, 'account:info', 'admin@192.168.18.106 登录成功！', 1688346106790, 1688346106790),
(2312, 'account:info', 'admin@192.168.18.106 登录成功！', 1688441639910, 1688441639910),
(2313, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688457599567, 1688457599567),
(2314, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688457608821, 1688457608821),
(2315, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688457620301, 1688457620301),
(2316, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688457717614, 1688457717614),
(2317, 'account:info', 'lvyitao@192.168.18.102 登录失败！ 错误信息:密码错误', 1688522046061, 1688522046061),
(2318, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688522050753, 1688522050753),
(2319, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688522126346, 1688522126346),
(2320, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688524892511, 1688524892511),
(2321, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688525178433, 1688525178433),
(2322, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688525913335, 1688525913335),
(2323, 'account:info', 'admin@172.3.200.9 登录成功！', 1688526503779, 1688526503779),
(2324, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688543517158, 1688543517158),
(2325, 'account:info', 'admin@192.168.18.102 登录成功！', 1688543585381, 1688543585381),
(2326, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688543639783, 1688543639783),
(2327, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688692322183, 1688692322183),
(2328, 'account:info', 'admin@192.168.18.102 登录成功！', 1688692340492, 1688692340492),
(2329, 'account:info', 'admin@172.24.0.1:12724 登录成功！', 1688719624142, 1688719624142),
(2330, 'account:info', 'admin@172.26.112.1:8065 登录成功！', 1688957660370, 1688957660370),
(2331, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1688967640292, 1688967640292),
(2332, 'account:info', 'admin@192.168.18.102 登录成功！', 1688967647184, 1688967647184),
(2333, 'account:info', 'admin@192.168.88.39 登录成功！', 1688969994148, 1688969994148),
(2334, 'account:info', 'lvyitao@192.168.18.102 登录成功！', 1689043989171, 1689043989171),
(2335, 'account:info', 'admin@192.168.18.102 登录成功！', 1689043999922, 1689043999922),
(2336, 'account:info', 'admin@192.168.18.102 登录成功！', 1689044109377, 1689044109377),
(2337, 'account:info', 'admin@172.0.0.21 登录成功！', 1689213546967, 1689213546967),
(2338, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689300718164, 1689300718164),
(2339, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689300934615, 1689300934615),
(2340, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689303774986, 1689303774986),
(2341, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689305259756, 1689305259756),
(2342, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689305428093, 1689305428093),
(2343, 'account:info', 'chenzhong@192.168.18.143 登录成功！', 1689305891118, 1689305891118),
(2344, 'account:info', 'admin@172.3.200.4 登录成功！', 1689477197822, 1689477197822),
(2345, 'account:info', 'admin@172.3.200.4 登录成功！', 1689942330652, 1689942330652),
(2346, 'account:info', 'admin@192.168.18.100 登录失败！ 错误信息:密码错误', 1690352750216, 1690352750216),
(2347, 'account:info', 'admin@192.168.18.100 登录失败！ 错误信息:密码错误', 1690352757790, 1690352757790),
(2348, 'account:info', 'admin@192.168.18.100 登录成功！', 1690352762499, 1690352762499);

-- --------------------------------------------------------

--
-- 表的结构 `user`
--

CREATE TABLE `user` (
  `id` bigint NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- 转储表的索引
--

--
-- 表的索引 `accounts`
--
ALTER TABLE `accounts`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `username` (`username`);

--
-- 表的索引 `accounts_token`
--
ALTER TABLE `accounts_token`
  ADD PRIMARY KEY (`token_key`),
  ADD KEY `uid` (`uid`),
  ADD KEY `expiration` (`expiration`);

--
-- 表的索引 `departments`
--
ALTER TABLE `departments`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `departments_accounts`
--
ALTER TABLE `departments_accounts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `account_id` (`account_id`),
  ADD KEY `department_id` (`department_id`);

--
-- 表的索引 `departments_jurisdictions`
--
ALTER TABLE `departments_jurisdictions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `department_id` (`department_id`),
  ADD KEY `jurisdiction_id` (`jurisdiction_id`);

--
-- 表的索引 `identitys`
--
ALTER TABLE `identitys`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`);

--
-- 表的索引 `identitys_accounts`
--
ALTER TABLE `identitys_accounts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `account_id` (`account_id`),
  ADD KEY `identity_id` (`identity_id`);

--
-- 表的索引 `identitys_jurisdictions`
--
ALTER TABLE `identitys_jurisdictions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `jurisdiction_id` (`jurisdiction_id`),
  ADD KEY `identity_id` (`identity_id`);

--
-- 表的索引 `jurisdictions`
--
ALTER TABLE `jurisdictions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`);

--
-- 表的索引 `locker`
--
ALTER TABLE `locker`
  ADD UNIQUE KEY `UQE_locker_guid` (`guid`),
  ADD UNIQUE KEY `UQE_locker_method` (`method`);

--
-- 表的索引 `logs`
--
ALTER TABLE `logs`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `accounts`
--
ALTER TABLE `accounts`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=199;

--
-- 使用表AUTO_INCREMENT `departments`
--
ALTER TABLE `departments`
  MODIFY `id` int NOT NULL AUTO_INCREMENT COMMENT '部门id', AUTO_INCREMENT=5273;

--
-- 使用表AUTO_INCREMENT `departments_accounts`
--
ALTER TABLE `departments_accounts`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=62;

--
-- 使用表AUTO_INCREMENT `departments_jurisdictions`
--
ALTER TABLE `departments_jurisdictions`
  MODIFY `id` int NOT NULL AUTO_INCREMENT COMMENT '权限流水id';

--
-- 使用表AUTO_INCREMENT `identitys`
--
ALTER TABLE `identitys`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=43;

--
-- 使用表AUTO_INCREMENT `identitys_accounts`
--
ALTER TABLE `identitys_accounts`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=239;

--
-- 使用表AUTO_INCREMENT `identitys_jurisdictions`
--
ALTER TABLE `identitys_jurisdictions`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=806;

--
-- 使用表AUTO_INCREMENT `jurisdictions`
--
ALTER TABLE `jurisdictions`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=183;

--
-- 使用表AUTO_INCREMENT `logs`
--
ALTER TABLE `logs`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2349;

--
-- 使用表AUTO_INCREMENT `user`
--
ALTER TABLE `user`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
