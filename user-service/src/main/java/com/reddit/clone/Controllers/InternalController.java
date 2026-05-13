package com.reddit.clone.Controllers;

import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserBlockRepository;
import com.reddit.clone.repository.UserRepository;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/internal")
public class InternalController {

    private final UserRepository userRepository;
    private final UserBlockRepository userBlockRepository;

    public InternalController(UserRepository userRepository,
                              UserBlockRepository userBlockRepository) {
        this.userRepository = userRepository;
        this.userBlockRepository = userBlockRepository;
    }

    @GetMapping("/users/{userId}/exists")
    public ResponseEntity<Map<String, Boolean>> userExists(@PathVariable String userId) {
        Optional<Long> id = parseId(userId);
        boolean exists = id.isPresent() && userRepository.existsById(id.get());
        return ResponseEntity.ok(Map.of("exists", exists));
    }

    @GetMapping("/users/{senderId}/blocked/{receiverId}")
    public ResponseEntity<Map<String, Boolean>> isBlocked(@PathVariable String senderId,
                                                          @PathVariable String receiverId) {
        Optional<User> sender = parseId(senderId).flatMap(userRepository::findById);
        Optional<User> receiver = parseId(receiverId).flatMap(userRepository::findById);
        boolean blocked = sender.isPresent()
                && receiver.isPresent()
                && userBlockRepository.existsByBlockerAndBlocked(sender.get(), receiver.get());

        return ResponseEntity.ok(Map.of("blocked", blocked));
    }

    private Optional<Long> parseId(String value) {
        try {
            return Optional.of(Long.parseLong(value));
        } catch (NumberFormatException e) {
            return Optional.empty();
        }
    }
}
