-- Channels table (server stores channel metadata, NOT messages)
CREATE TABLE channels (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Channel members (who belongs to which channel)
CREATE TABLE channel_members (
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

-- Indexes
CREATE INDEX idx_channels_name ON channels(name);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);

-- Trigger for updated_at
CREATE TRIGGER update_channels_updated_at
    BEFORE UPDATE ON channels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();