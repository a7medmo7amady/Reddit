package com.reddit.clone.repository;

import com.reddit.clone.model.ModerationLog;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ModerationLogRepository extends MongoRepository<ModerationLog, String> {
}
