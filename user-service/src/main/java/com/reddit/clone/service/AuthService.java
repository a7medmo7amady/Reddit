package com.reddit.clone.service;

import com.reddit.clone.exception.InvalidCredentialsException;
import com.reddit.clone.exception.OAuthLoginRequiredException;
import com.reddit.clone.exception.UserAlreadyExistsException;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.factory.UserFactory;
import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserRepository;
import com.reddit.clone.security.PasswordService;
import org.springframework.stereotype.Service;

@Service
public class AuthService {

    private final UserRepository  userRepository;
    private final UserFactory     userFactory;
    private final PasswordService passwordService;

    public AuthService(UserRepository userRepository,
                       UserFactory userFactory,
                       PasswordService passwordService) {
        this.userRepository  = userRepository;
        this.userFactory     = userFactory;
        this.passwordService = passwordService;
    }

    public User signup(String username, String email, String password) {
        if (password == null || password.isBlank()) {
            throw new InvalidCredentialsException("Password is required for standard signup");
        }
        validateNewUser(username, email);
        User user = userFactory.createUser(username, email, passwordService.hash(password));
        return userRepository.save(user);
    }

    public User oauthSignup(String username, String email, OAuthProvider provider) {
        validateNewUser(username, email);
        User user = userFactory.createOAuthUser(username, email, provider);
        return userRepository.save(user);
    }

    public User findOrCreateOAuthUser(String email, String name, String avatar, OAuthProvider provider) {
        return userRepository.findByEmail(email)
                .map(user -> {
                    if (user.isBanned()) {
                        throw new SecurityException("User is banned");
                    }
                    if (user.getOauthProvider() == null) {
                        user.setOauthProvider(provider);
                    }
                    if (user.getDisplayName() == null || user.getDisplayName().isBlank()) {
                        user.setDisplayName(name);
                    }
                    if (user.getAvatar() == null || user.getAvatar().isBlank()) {
                        user.setAvatar(avatar);
                    }
                    return userRepository.save(user);
                })
                .orElseGet(() -> {
                    String username = uniqueUsername(email, name);
                    User user = userFactory.createOAuthUser(username, email, provider);
                    user.setDisplayName(name);
                    user.setAvatar(avatar);
                    return userRepository.save(user);
                });
    }

    public User findById(Long id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("User not found"));
    }

    public User login(String email, String password) {
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new UserNotFoundException("Invalid email or password"));

        if (user.isBanned()) {
            throw new SecurityException("User is banned");
        }

        if (user.getOauthProvider() != null) {
            throw new OAuthLoginRequiredException(user.getOauthProvider().toString());
        }

        if (password == null || password.isBlank()) {
            throw new InvalidCredentialsException("Password is required");
        }
        if (!passwordService.verify(password, user.getPassword())) {
            throw new UserNotFoundException("Invalid email or password");
        }

        return user;
    }

    private void validateNewUser(String username, String email) {
        if (userRepository.existsByUsername(username)) {
            throw new UserAlreadyExistsException("Username already taken: " + username);
        }
        if (userRepository.existsByEmail(email)) {
            throw new UserAlreadyExistsException("Email already registered: " + email);
        }
    }

    private String uniqueUsername(String email, String name) {
        String source = name != null && !name.isBlank() ? name : email.substring(0, email.indexOf('@'));
        String base = source.toLowerCase()
                .replaceAll("[^a-z0-9_]+", "_")
                .replaceAll("^_+|_+$", "");
        if (base.isBlank()) {
            base = "user";
        }

        String candidate = base;
        int suffix = 1;
        while (userRepository.existsByUsername(candidate)) {
            candidate = base + "_" + suffix++;
        }
        return candidate;
    }
}
