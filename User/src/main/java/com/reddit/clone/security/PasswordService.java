package com.reddit.clone.security;

import org.bouncycastle.crypto.generators.Argon2BytesGenerator;
import org.bouncycastle.crypto.params.Argon2Parameters;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Base64;

@Component
public class PasswordService {

    private static final int          ITERATIONS    = 3;
    private static final int          MEMORY_KB     = 16 * 1024;
    private static final int          PARALLELISM   = 4;
    private static final int          HASH_LENGTH   = 32;
    private static final int          SALT_LENGTH   = 16;
    private static final SecureRandom SECURE_RANDOM = new SecureRandom();

    public String hash(String password) {
        byte[] salt = new byte[SALT_LENGTH];
        SECURE_RANDOM.nextBytes(salt);
        byte[] hash = new byte[HASH_LENGTH];

        buildAndRun(password.getBytes(StandardCharsets.UTF_8), salt, hash);

        return ITERATIONS + ":" + MEMORY_KB + ":" + PARALLELISM + ":"
             + Base64.getEncoder().encodeToString(salt) + ":"
             + Base64.getEncoder().encodeToString(hash);
    }

    public boolean verify(String rawPassword, String encoded) {
        if (encoded == null) return false;
        String[] parts = encoded.split(":");
        if (parts.length != 5) return false;

        try {
            int    iterations  = Integer.parseInt(parts[0]);
            int    memoryKB    = Integer.parseInt(parts[1]);
            int    parallelism = Integer.parseInt(parts[2]);
            byte[] salt         = Base64.getDecoder().decode(parts[3]);
            byte[] expectedHash = Base64.getDecoder().decode(parts[4]);
            byte[] actualHash   = new byte[HASH_LENGTH];

            Argon2Parameters parameters = new Argon2Parameters.Builder(Argon2Parameters.ARGON2_id)
                    .withIterations(iterations)
                    .withMemoryAsKB(memoryKB)
                    .withParallelism(parallelism)
                    .withSalt(salt)
                    .build();

            Argon2BytesGenerator generator = new Argon2BytesGenerator();
            generator.init(parameters);
            generator.generateBytes(rawPassword.getBytes(StandardCharsets.UTF_8), actualHash);

            return MessageDigest.isEqual(expectedHash, actualHash);
        } catch (IllegalArgumentException e) {
            return false;
        }
    }

    private void buildAndRun(byte[] password, byte[] salt, byte[] out) {
        Argon2Parameters parameters = new Argon2Parameters.Builder(Argon2Parameters.ARGON2_id)
                .withIterations(ITERATIONS)
                .withMemoryAsKB(MEMORY_KB)
                .withParallelism(PARALLELISM)
                .withSalt(salt)
                .build();

        Argon2BytesGenerator generator = new Argon2BytesGenerator();
        generator.init(parameters);
        generator.generateBytes(password, out);
    }
}
