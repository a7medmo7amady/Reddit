package com.reddit.clone.kafka;

import com.reddit.clone.event.CommunityBanEvent;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class BanEventPublisher {

    private static final String TOPIC = "community.ban";

    private final KafkaTemplate<String, CommunityBanEvent> kafkaTemplate;

    public BanEventPublisher(KafkaTemplate<String, CommunityBanEvent> kafkaTemplate) {
        this.kafkaTemplate = kafkaTemplate;
    }

    public void publish(CommunityBanEvent event) {
        String key = event.userId() + ":" + event.community();
        kafkaTemplate.send(TOPIC, key, event);
    }
}
