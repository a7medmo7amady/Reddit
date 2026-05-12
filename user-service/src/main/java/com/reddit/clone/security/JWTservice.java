package com.reddit.clone.security;

import com.reddit.clone.model.User;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Date;
import java.util.Locale;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class JWTservice {

    private static final Pattern DURATION_PATTERN = Pattern.compile("^(\\d+)([smhd])$");

    @Value("${jwt.secret:dev-only-secret-key-change-in-production-must-be-32chars!!}")
    private String secret;

    @Value("${jwt.expires-in:15m}")
    private String expiresIn;

    public String generateAccessToken(User user) {
        return Jwts.builder()
                .subject(String.valueOf(user.getId()))
                .claim("username", user.getUsername())
                .claim("role", user.getRole().name())
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + expirationMillis()))
                .signWith(signingKey())
                .compact();
    }

    public Long extractUserId(String token) {
        return Long.parseLong(parse(token).getSubject());
    }

    private Claims parse(String token) {
        return Jwts.parser()
                .verifyWith(signingKey())
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    private SecretKey signingKey() {
        return Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8));
    }

    private long expirationMillis() {
        Matcher matcher = DURATION_PATTERN.matcher(expiresIn.trim().toLowerCase(Locale.ROOT));
        if (!matcher.matches()) {
            throw new IllegalArgumentException("Invalid jwt.expires-in value: " + expiresIn);
        }

        long amount = Long.parseLong(matcher.group(1));
        return switch (matcher.group(2)) {
            case "s" -> Duration.ofSeconds(amount).toMillis();
            case "m" -> Duration.ofMinutes(amount).toMillis();
            case "h" -> Duration.ofHours(amount).toMillis();
            case "d" -> Duration.ofDays(amount).toMillis();
            default -> throw new IllegalArgumentException("Invalid jwt.expires-in unit: " + expiresIn);
        };
    }
}
