/*
 Navicat Premium Data Transfer

 Source Server         : 10.64.1.87_33306
 Source Server Type    : MySQL
 Source Server Version : 50738 (5.7.38)
 Source Host           : 10.64.1.87:33306
 Source Schema         : database-api

 Target Server Type    : MySQL
 Target Server Version : 50738 (5.7.38)
 File Encoding         : 65001

 Date: 26/04/2023 10:34:59
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for api
-- ----------------------------
DROP TABLE IF EXISTS `api`;
CREATE TABLE `api`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '分组ID',
  `api_mark` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'API标识',
  `project_id` bigint(20) UNSIGNED NOT NULL COMMENT '项目ID',
  `user_id` bigint(20) UNSIGNED NOT NULL COMMENT '接口负责人ID',
  `source_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据源ID',
  `api_name` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '接口名称',
  `payload` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '接口负载（sql语句等信息）',
  `action` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作属性 0:创建 1:更新 2:查询 3:删除',
  `state` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布状态 0:开发中 1:已上线 2:已下线',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`, `api_mark`) USING BTREE,
  UNIQUE INDEX `api_mark`(`api_mark`) USING BTREE,
  INDEX `group_id`(`group_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of api
-- ----------------------------
INSERT INTO `api` VALUES (1, 1, '123', 2, 2, 3, '测试', 'select * from sys_user where user_name=@username', 2, 1, '2023-02-01 16:37:14', '2023-03-16 14:41:10');
INSERT INTO `api` VALUES (4, 1, '6daa8331-58c4-453f-9a0a-d080637f77bc', 2, 2, 3, '获取项目列表', 'select * from sys_user where user_name = @username', 2, 1, '2023-03-16 10:45:07', '2023-03-17 14:08:47');
INSERT INTO `api` VALUES (5, 4, '5b1aa191-98c9-45fa-9360-0e14a8b93b13', 2, 2, 4, '测试API', '', 4, 0, '2023-03-16 14:03:42', '2023-03-16 14:03:42');

-- ----------------------------
-- Table structure for api_group
-- ----------------------------
DROP TABLE IF EXISTS `api_group`;
CREATE TABLE `api_group`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '项目ID',
  `group_name` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '分组名称',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `project_id`(`project_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of api_group
-- ----------------------------
INSERT INTO `api_group` VALUES (1, 2, '测试分组', '2023-02-02 10:14:08', '2023-03-14 10:26:02');
INSERT INTO `api_group` VALUES (2, 6, '测试分组', '2023-02-08 09:59:22', '2023-02-08 09:59:22');
INSERT INTO `api_group` VALUES (3, 3, '测试分组', '2023-03-03 13:47:47', '2023-03-03 13:47:47');
INSERT INTO `api_group` VALUES (4, 2, '测试分组3', '2023-03-06 13:19:47', '2023-03-13 14:45:53');
INSERT INTO `api_group` VALUES (5, 2, '测试分组5', '2023-03-06 13:24:42', '2023-03-13 14:45:57');
INSERT INTO `api_group` VALUES (6, 2, '测试分组', '2023-03-06 13:37:54', '2023-03-06 13:37:54');

-- ----------------------------
-- Table structure for project
-- ----------------------------
DROP TABLE IF EXISTS `project`;
CREATE TABLE `project`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `project_name` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '项目名称',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of project
-- ----------------------------
INSERT INTO `project` VALUES (2, 2, '测试项目', '2023-02-01 10:21:27', '2023-03-02 13:26:18');
INSERT INTO `project` VALUES (3, 3, '111', '2023-02-02 11:13:26', '2023-02-02 11:13:26');
INSERT INTO `project` VALUES (4, 2, 'qq', '2023-03-06 10:32:02', '2023-03-06 10:51:24');

-- ----------------------------
-- Table structure for source
-- ----------------------------
DROP TABLE IF EXISTS `source`;
CREATE TABLE `source`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `source_mark` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据源标识',
  `project_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属项目ID',
  `user_id` bigint(20) NOT NULL COMMENT '所属用户ID',
  `source_name` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据源名称',
  `device` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库驱动',
  `ip` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库地址',
  `port` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据库端口',
  `database` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库名称',
  `charset` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库编码',
  `user` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库用户名',
  `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库密码',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `project_id`(`project_id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of source
-- ----------------------------
INSERT INTO `source` VALUES (3, '123', 2, 2, 'AI数据源', 'mysql', '10.64.1.87', 33306, 'ai_online', 'utf8mb4', 'root', '1qaz@WSX', '2023-02-01 10:35:49', '2023-03-13 15:08:15');
INSERT INTO `source` VALUES (4, '456', 2, 2, 'AI数据源23', 'mysql', '10.64.1.87', 33306, 'ai_online', 'utf8mb4', 'root', '1qaz@WSX', '2023-02-01 10:42:45', '2023-03-16 15:27:05');
INSERT INTO `source` VALUES (6, '789', 6, 3, 'Ai数据源2', 'mysql', '10.64.1.87', 33306, 'ai_online', 'utf8mb4', 'root', '1qaz@WSX', '2023-02-06 15:41:22', '2023-02-06 16:33:42');
INSERT INTO `source` VALUES (7, '011', 6, 3, 'ide数据源', 'mysql', '10.64.1.87', 33306, 'unifly-ide-new', 'utf8mb4', 'root', '1qaz@WSX', '2023-02-07 15:38:57', '2023-02-07 15:38:57');
INSERT INTO `source` VALUES (8, '5ad59dea-c5e4-450d-ae0d-0352ebcdc103', 2, 2, 'test', 'mysql', '10.64.1.87', 33306, 'ai_online', 'utf8mb4', 'root', '1qaz@WSX', '2023-03-16 16:24:45', '2023-03-16 16:24:45');

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `nickname` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '显示名称',
  `password` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '密码',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES (2, 'gzxy', 'gzxy', '$2a$10$EajzmBv1gQuTnqQ69ItVbONXw/7D4eliyV2efVRprshrMhcXlDuWy', '2023-01-29 16:04:23', '2023-01-29 16:04:23');
INSERT INTO `user` VALUES (3, 'admin', '幸运', '$2a$10$l5PPlALJ2SH1gR.btwLeau7sYBJkQNdJCujYNw1yLRuPNT/RLKTa2', '2023-02-01 08:54:36', '2023-02-01 08:54:36');

SET FOREIGN_KEY_CHECKS = 1;
