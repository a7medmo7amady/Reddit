package com.reddit.clone.service;

import com.reddit.clone.exception.UserAlreadyExistsException;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.factory.UserFactory;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserRepository;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UserService {

    private final UserRepository userRepository;
    private final UserFactory userFactory;

    public UserService(UserRepository userRepository, UserFactory userFactory) {
        this.userRepository = userRepository;
        this.userFactory = userFactory;
    }

    public User register(String username, String email, String password) {
        if (userRepository.existsByUsername(username)) {
            throw new UserAlreadyExistsException("Username already taken: " + username);
        }
        if (userRepository.existsByEmail(email)) {
            throw new UserAlreadyExistsException("Email already registered: " + email);
        }
        User user = userFactory.createUser(username, email, password);
        return userRepository.save(user);
    }

    public User findById(Long id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("User not found with id: " + id));
    }

    public User findByUsername(String username) {
        return userRepository.findByUsername(username)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + username));
    }

    public List<User> findAll() {
        return userRepository.findAll();
    }

    public User updateUser(Long id, String username, String email, String name, String bio, String avatar) {
        User user = findById(id);
        if (username != null && !username.equals(user.getUsername())) {
            if (userRepository.existsByUsername(username)) {
                throw new UserAlreadyExistsException("Username already taken: " + username);
            }
            user.setUsername(username);
        }
        if (email != null && !email.equals(user.getEmail())) {
            if (userRepository.existsByEmail(email)) {
                throw new UserAlreadyExistsException("Email already registered: " + email);
            }
            user.setEmail(email);
        }
        if (name != null)   user.setDisplayName(name);
        if (bio != null)    user.setBio(bio);
        if (avatar != null) user.setAvatar(avatar);
        return userRepository.save(user);
    }

    public void deleteUser(Long id) {
        if (!userRepository.existsById(id)) {
            throw new UserNotFoundException("User not found with id: " + id);
        }
        userRepository.deleteById(id);
    }
}
