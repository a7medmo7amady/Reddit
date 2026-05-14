package com.reddit.clone.dto;

import java.time.LocalDateTime;

public record PublicProfileResponse(
        Long id,
        String username,
        String displayName,
        String bio,
        String avatar,
        String banner,
        int karma,
        LocalDateTime createdAt
) {}
