/*
 Navicat Premium Data Transfer

 Source Server         : 172.0.0.41
 Source Server Type    : MySQL
 Source Server Version : 80026
 Source Host           : 172.0.0.41:3306
 Source Schema         : dev_account_center

 Target Server Type    : MySQL
 Target Server Version : 80026
 File Encoding         : 65001

 Date: 19/07/2023 09:13:19
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for accounts
-- ----------------------------
DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `salt` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `can_delete` tinyint(1) NULL DEFAULT NULL,
  `created_at` bigint(0) NULL DEFAULT NULL,
  `updated_at` bigint(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 199 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of accounts
-- ----------------------------
INSERT INTO `accounts` VALUES (1, 'superadmin', '超级管理员', 'd445a3c28088b38d9d7d037201255109', 'aa', 'sUOHoF', 0, 0, 1667878497);
INSERT INTO `accounts` VALUES (2, 'admin', '管理员', '625ecf15c5f267ed3f3621e291b48150', 'aa', 'MTfldn', 0, 0, 1681106639);

-- ----------------------------
-- Table structure for accounts_token
-- ----------------------------
DROP TABLE IF EXISTS `accounts_token`;
CREATE TABLE `accounts_token`  (
  `token_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `uid` int(0) NOT NULL,
  `expiration` int(0) NOT NULL,
  `updated_at` int(0) NOT NULL,
  `created_at` int(0) NOT NULL,
  PRIMARY KEY (`token_key`) USING BTREE,
  INDEX `uid`(`uid`) USING BTREE,
  INDEX `expiration`(`expiration`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of accounts_token
-- ----------------------------
INSERT INTO `accounts_token` VALUES ('d1d62a652a3f4978089556e83beaecdc', 2, 1691457709, 1689729709, 1689729709);

-- ----------------------------
-- Table structure for departments
-- ----------------------------
DROP TABLE IF EXISTS `departments`;
CREATE TABLE `departments`  (
  `id` int(0) NOT NULL AUTO_INCREMENT COMMENT '部门id',
  `fid` bigint(0) NULL DEFAULT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '部门查找key',
  `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '部门名称（显示用）',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '备注',
  `updated_at` bigint(0) NOT NULL COMMENT '更新时间',
  `created_at` bigint(0) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5273 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for departments_accounts
-- ----------------------------
DROP TABLE IF EXISTS `departments_accounts`;
CREATE TABLE `departments_accounts`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `account_id` int(0) NOT NULL,
  `department_id` int(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `account_id`(`account_id`) USING BTREE,
  INDEX `department_id`(`department_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 62 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for departments_jurisdictions
-- ----------------------------
DROP TABLE IF EXISTS `departments_jurisdictions`;
CREATE TABLE `departments_jurisdictions`  (
  `id` int(0) NOT NULL AUTO_INCREMENT COMMENT '权限流水id',
  `department_id` int(0) NOT NULL COMMENT '所属部门',
  `jurisdiction_id` int(0) NOT NULL COMMENT '权限id（索引）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `department_id`(`department_id`) USING BTREE,
  INDEX `jurisdiction_id`(`jurisdiction_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for identitys
-- ----------------------------
DROP TABLE IF EXISTS `identitys`;
CREATE TABLE `identitys`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `is_system` int(0) NOT NULL DEFAULT 0,
  `updated_at` bigint(0) NOT NULL,
  `created_at` bigint(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `code`(`code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 43 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of identitys
-- ----------------------------
INSERT INTO `identitys` VALUES (1, 'superadmin', '超级管理员', '', 1, 0, 0);

-- ----------------------------
-- Table structure for identitys_accounts
-- ----------------------------
DROP TABLE IF EXISTS `identitys_accounts`;
CREATE TABLE `identitys_accounts`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `identity_id` int(0) NOT NULL,
  `account_id` int(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `account_id`(`account_id`) USING BTREE,
  INDEX `identity_id`(`identity_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 239 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of identitys_accounts
-- ----------------------------
INSERT INTO `identitys_accounts` VALUES (1, 1, 1);
INSERT INTO `identitys_accounts` VALUES (2, 1, 2);

-- ----------------------------
-- Table structure for identitys_jurisdictions
-- ----------------------------
DROP TABLE IF EXISTS `identitys_jurisdictions`;
CREATE TABLE `identitys_jurisdictions`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `identity_id` int(0) NOT NULL COMMENT '所属身份',
  `jurisdiction_id` int(0) NOT NULL COMMENT '权限id（索引）',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `jurisdiction_id`(`jurisdiction_id`) USING BTREE,
  INDEX `identity_id`(`identity_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 806 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of identitys_jurisdictions
-- ----------------------------
INSERT INTO `identitys_jurisdictions` VALUES (1, 1, 1);

-- ----------------------------
-- Table structure for jurisdictions
-- ----------------------------
DROP TABLE IF EXISTS `jurisdictions`;
CREATE TABLE `jurisdictions`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `fid` bigint(0) NULL DEFAULT 0,
  `is_system` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `code`(`code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 183 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of jurisdictions
-- ----------------------------
INSERT INTO `jurisdictions` VALUES (1, 'superadmin', '超级管理员', 0, 1);

-- ----------------------------
-- Table structure for locker
-- ----------------------------
DROP TABLE IF EXISTS `locker`;
CREATE TABLE `locker`  (
  `guid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `method` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `created_at` bigint(0) NULL DEFAULT NULL,
  UNIQUE INDEX `UQE_locker_guid`(`guid`) USING BTREE,
  UNIQUE INDEX `UQE_locker_method`(`method`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for logs
-- ----------------------------
DROP TABLE IF EXISTS `logs`;
CREATE TABLE `logs`  (
  `id` bigint(0) NOT NULL AUTO_INCREMENT,
  `type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `created_at` bigint(0) NULL DEFAULT NULL,
  `updated_at` bigint(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2345 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of logs
-- ----------------------------
INSERT INTO `logs` VALUES (2345, 'account:info', 'admin@ws://172.0.0.41:5000/websocket/jsonrpc 登录成功！', 1689729709912, 1689729709912);

SET FOREIGN_KEY_CHECKS = 1;
