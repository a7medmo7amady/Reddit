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

    public User findById(Long id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("User not found"));
    }

    public User login(String email, String password) {
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new UserNotFoundException("Invalid email or password"));

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
}
