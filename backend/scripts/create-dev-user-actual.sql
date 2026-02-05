-- Create dev admin user and API key in terraform_registry database
-- API Key: dev_qHlTX4JvjK1yVUgRukLlgiwFQmFOiHdEhHYVJNfhNXc
-- Key Prefix: dev_qHlTX4
-- Hash: $2a$10$.R.qK7nGshGkADopU53nk.WKX6pHx1ync3Vg4w6aITQCO9fCioWza

DO $$
DECLARE
    v_user_id uuid;
BEGIN
    -- Insert admin user
    INSERT INTO users (email, name, oidc_sub)
    VALUES (
        'admin@dev.local',
        'Dev Admin',
        'dev-admin-oidc-sub'
    )
    ON CONFLICT (email) DO NOTHING;
    
    -- Get the user ID
    SELECT id INTO v_user_id FROM users WHERE email = 'admin@dev.local';
    RAISE NOTICE 'User ID: %', v_user_id;
    
    -- Delete existing API key if it exists
    DELETE FROM api_keys WHERE key_prefix = 'dev_qHlTX4';
    
    -- Create new API key
    INSERT INTO api_keys (user_id, name, key_hash, key_prefix, scopes, created_at)
    VALUES (
        v_user_id,
        'Development API Key',
        '$2a$10$.R.qK7nGshGkADopU53nk.WKX6pHx1ync3Vg4w6aITQCO9fCioWza',
        'dev_qHlTX4',
        '["admin"]'::jsonb,
        NOW()
    );
    
    RAISE NOTICE 'API key created successfully';
END $$;

-- Verify
SELECT id, email, name FROM users WHERE email = 'admin@dev.local';
SELECT id, user_id, name, key_prefix, scopes FROM api_keys WHERE key_prefix = 'dev_qHlTX4';
