-- Polykeys configuration example
-- Format: { "alias", "deviceID", "layout" }

mappings = {
    -- Map devices to keyboard layouts
    -- The alias is for display, deviceID is the actual VID:PID
    { "Corne", "4653:0004", "US International" },
    { "Lily58", "1209:bb58", "US" },
    { "Logitech K380", "046d:c52b", "US" },

    -- Fallback when no device matches (use any deviceID)
    { "System Default", "system_default", "French AZERTY" },
}

-- Old 2-element format is also supported for compatibility:
-- { "deviceID", "layout" }
