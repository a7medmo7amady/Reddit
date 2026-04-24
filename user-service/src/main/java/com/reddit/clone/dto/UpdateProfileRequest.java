package com.reddit.clone.dto;

import jakarta.validation.constraints.Size;

import java.util.List;

public record UpdateProfileRequest(
        @Size(max = 50) String displayName,
        @Size(max = 500) String bio,
        String avatar,
        String banner,
        @Size(max = 3) List<String> links
) {}
