-- Notification System Database Migration
-- Run this script to add notification functionality to the existing database

-- Create notifications table
CREATE TABLE `notifications` (
  `id_notification` varchar(36) NOT NULL,
  `target_member_id` varchar(36) NOT NULL,
  `from_member_id` varchar(36) NOT NULL,
  `type` enum('status_change_request', 'general') NOT NULL,
  `title` varchar(255) NOT NULL,
  `message` text NOT NULL,
  `metadata` json DEFAULT NULL,
  `read_at` datetime NULL,
  `pending` boolean DEFAULT TRUE,
  `accepted` boolean NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id_notification`),
  KEY `idx_target_member` (`target_member_id`),
  KEY `idx_created_at` (`created_at` DESC),
  KEY `idx_target_unread` (`target_member_id`, `read_at`),
  
  CONSTRAINT `notifications_target_fk` FOREIGN KEY (`target_member_id`) REFERENCES `member` (`id_member`) ON DELETE CASCADE,
  CONSTRAINT `notifications_from_fk` FOREIGN KEY (`from_member_id`) REFERENCES `member` (`id_member`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Create status_change_requests table
CREATE TABLE `status_change_requests` (
  `id_request` varchar(36) NOT NULL,
  `notification_id` varchar(36) NOT NULL,
  `target_member_id` varchar(36) NOT NULL,
  `requested_by_member_id` varchar(36) NOT NULL,
  `from_status` enum('anggota','bph','alb','dpo','bp') NOT NULL,
  `to_status` enum('anggota','bph','alb','dpo','bp') NOT NULL,
  `status` enum('pending', 'accepted', 'rejected') DEFAULT 'pending',
  `processed_at` datetime NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id_request`),
  KEY `idx_target_status` (`target_member_id`, `status`),
  KEY `idx_notification` (`notification_id`),
  
  CONSTRAINT `status_requests_notification_fk` FOREIGN KEY (`notification_id`) REFERENCES `notifications` (`id_notification`) ON DELETE CASCADE,
  CONSTRAINT `status_requests_target_fk` FOREIGN KEY (`target_member_id`) REFERENCES `member` (`id_member`) ON DELETE CASCADE,
  CONSTRAINT `status_requests_requester_fk` FOREIGN KEY (`requested_by_member_id`) REFERENCES `member` (`id_member`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Add performance indexes
CREATE INDEX idx_notifications_target_pending ON notifications(target_member_id, pending);
CREATE INDEX idx_notifications_type_created ON notifications(type, created_at DESC);
CREATE INDEX idx_status_requests_target_status ON status_change_requests(target_member_id, status);

-- Sample data (optional, for testing)
-- Uncomment the following lines if you want to add sample notification data

-- INSERT INTO notifications (id_notification, target_member_id, from_member_id, type, title, message, metadata, pending, created_at, updated_at)
-- VALUES 
-- ('550e8400-e29b-41d4-a716-446655440000', 'target_member_id_here', 'from_member_id_here', 'status_change_request', 'Status Change Request', 'Sample notification message', '{"request_id": "req123", "from_status": "bph", "to_status": "dpo"}', true, NOW(), NOW());

-- INSERT INTO status_change_requests (id_request, notification_id, target_member_id, requested_by_member_id, from_status, to_status, status, created_at)
-- VALUES 
-- ('req123', '550e8400-e29b-41d4-a716-446655440000', 'target_member_id_here', 'from_member_id_here', 'bph', 'dpo', 'pending', NOW());

-- Verify the tables were created successfully
SHOW TABLES LIKE '%notification%';
SHOW TABLES LIKE '%status_change%';

-- Show table structures
DESCRIBE notifications;
DESCRIBE status_change_requests;
