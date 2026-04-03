-- Add thinking/reasoning column to messages for models that support extended thinking
ALTER TABLE messages ADD COLUMN IF NOT EXISTS thinking TEXT;
