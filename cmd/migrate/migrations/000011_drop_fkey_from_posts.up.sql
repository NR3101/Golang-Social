-- Drop the foreign key constraint from posts table
ALTER TABLE posts
    DROP CONSTRAINT posts_user_id_fkey;