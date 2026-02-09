-- Add test links with clicks for your user
-- Replace 'YOUR_FIREBASE_UID' with your actual Firebase user ID

-- First, let's check your user ID
SELECT id, email FROM users;

-- Insert test links (replace the user_id with yours from the query above)
INSERT INTO links (id, original_url, short_code, created_at, expires_at, clicks, user_id) 
VALUES 
  (1001, 'https://summer-tech-conference.com/tickets/early-bird-registration', 'sum-conf-24', NOW() - INTERVAL '5 days', NULL, 3456, 'YOUR_FIREBASE_UID'),
  (1002, 'https://portfolio-alex-design.webflow.io/projects/ux-case-study', 'port-24', NOW() - INTERVAL '3 days', NULL, 892, 'YOUR_FIREBASE_UID'),
  (1003, 'https://bento.me/alexthompson-designs', 'bento-box', NOW() - INTERVAL '1 day', NULL, 234, 'YOUR_FIREBASE_UID')
ON CONFLICT (id) DO NOTHING;
