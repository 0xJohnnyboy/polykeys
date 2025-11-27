-- Polykeys configuration example

mappings = {
    -- Map device names to keyboard layouts
    { "Corne", "US International" },
    { "Lily58", "US" },

    -- You can also use device IDs (use "polykeys logs -f" to find them)
    { "046d:c52b", "US" },

    -- Fallback when no device matches
    { "system_default", "French" },
}
