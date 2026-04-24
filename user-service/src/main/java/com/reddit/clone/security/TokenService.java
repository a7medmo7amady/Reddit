package com.reddit.clone.security;

import com.reddit.clone.exception.InvalidCredentialsException;
import com.reddit.clone.model.RefreshToken;
import com.reddit.clone.repository.RefreshTokenRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.time.LocalDateTime;
import java.util.Base64;

@Service
public class TokenService {

    private static final int REFRESH_EXPIRY_DAYS = 7;

    private final RefreshTokenRepository refreshTokenRepository;

    public TokenService(RefreshTokenRepository refreshTokenRepository) {
        this.refreshTokenRepository = refreshTokenRepository;
    }

    public record RotationResult(Long userId, String newToken) {}

    public String create(Long userId) {
        String raw = secureRandom();
        refreshTokenRepository.save(
                new RefreshToken(hash(raw), userId, LocalDateTime.now().plusDays(REFRESH_EXPIRY_DAYS)));
        return raw;
    }

    @Transactional
    public RotationResult rotate(String raw) {
        RefreshToken existing = loadValid(raw);
        existing.setRevoked(true);
        refreshTokenRepository.save(existing);
        return new RotationResult(existing.getUserId(), create(existing.getUserId()));
    }

    @Transactional
    public void revoke(String raw) {
        refreshTokenRepository.findByTokenHash(hash(raw)).ifPresent(rt -> {
            rt.setRevoked(true);
            refreshTokenRepository.save(rt);
        });
    }

    @Transactional
    public void revokeAllForUser(Long userId) {
        refreshTokenRepository.deleteAllByUserId(userId);
    }

    private RefreshToken loadValid(String raw) {
        RefreshToken rt = refreshTokenRepository.findByTokenHash(hash(raw))
                .orElseThrow(() -> new InvalidCredentialsException("Invalid refresh token"));
        if (rt.isRevoked() || rt.getExpiresAt().isBefore(LocalDateTime.now())) {
            throw new InvalidCredentialsException("Refresh token expired or revoked");
        }
        return rt;
    }

    private String secureRandom() {
        byte[] bytes = new byte[32];
        new SecureRandom().nextBytes(bytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }

    private String hash(String raw) {
        try {
            byte[] digest = MessageDigest.getInstance("SHA-256")
                    .digest(raw.getBytes(StandardCharsets.UTF_8));
            return Base64.getEncoder().encodeToString(digest);
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException("SHA-256 unavailable", e);
        }
    }
}
