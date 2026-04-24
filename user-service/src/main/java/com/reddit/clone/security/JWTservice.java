package com.reddit.clone.security;

import com.reddit.clone.model.User;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.Date;

@Service
public class JWTservice {

    private static final long ACCESS_TOKEN_EXPIRY_MS = 15 * 60 * 1000L;

    @Value("${jwt.secret:dev-only-secret-key-change-in-production-must-be-32chars!!}")
    private String secret;

    public String generateAccessToken(User user) {
        return Jwts.builder()
                .subject(String.valueOf(user.getId()))
                .claim("username", user.getUsername())
                .claim("role", user.getRole().name())
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + ACCESS_TOKEN_EXPIRY_MS))
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
}
