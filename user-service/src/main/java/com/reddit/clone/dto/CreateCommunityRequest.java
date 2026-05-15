package com.reddit.clone.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

public record CreateCommunityRequest(
        @NotBlank
        @Size(min = 3, max = 21)
        @Pattern(regexp = "^[a-zA-Z0-9_]+$", message = "Community name can only contain letters, numbers, and underscores")
        String name,

        @Size(max = 500)
        String description
) {}
