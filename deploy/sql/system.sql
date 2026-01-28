CREATE TABLE `system_monitor_config` (
                                         `id` int NOT NULL AUTO_INCREMENT,
                                         `is_start` tinyint(1) DEFAULT NULL,
                                         `cpu_limit` float DEFAULT NULL,
                                         `disk_limit` float DEFAULT NULL,
                                         `men_limit` float DEFAULT NULL,
                                         `net_send_limit` float DEFAULT NULL,
                                         `net_recv_limit` float DEFAULT NULL,
                                         `notify_type` tinyint(4) DEFAULT NULL,
                                         `email` varchar(255) DEFAULT NULL,
                                         `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                         `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                         PRIMARY KEY (`id`)
) ENGINE=MyISAM  AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

CREATE TABLE `system_monitor_warning` (
                                          `id` int NOT NULL AUTO_INCREMENT,
                                          `state_type` varchar(255) NOT NULL,
                                          `limit_value` float NOT NULL,
                                          `state_value` float NOT NULL,
                                          `occurrence` timestamp NULL DEFAULT NULL,
                                          `is_notify` tinyint DEFAULT '0',
                                          `day` int NOT NULL,
                                          `create_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
                                          `update_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
                                          PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb3;
