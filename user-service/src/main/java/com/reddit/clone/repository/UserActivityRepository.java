package com.reddit.clone.repository;

import com.reddit.clone.model.UserActivity;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.util.Optional;

public interface UserActivityRepository extends MongoRepository<UserActivity, String> {
    Optional<UserActivity> findByUserId(Long userId);
}
