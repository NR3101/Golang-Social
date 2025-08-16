-- Re-add the foreign key constraint to posts table
ALTER TABLE posts
    ADD CONSTRAINT posts_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;