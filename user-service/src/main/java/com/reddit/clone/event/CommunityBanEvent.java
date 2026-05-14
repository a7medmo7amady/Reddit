package com.reddit.clone.event;

public record CommunityBanEvent(
        Long   userId,
        String username,
        String community,
        String action,   
        String reason,
        String occurredAt
) {}
