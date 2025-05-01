plugins {
    id("java")
    id("application")  // Добавление плагина application
}

group = "com.okolelov"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

application {
    mainClass.set("com.okolelov.ClickThroughRateProcessor")  // Правильный синтаксис для Kotlin DSL
}

dependencies {
    // Kafka Streams
    implementation("org.apache.kafka:kafka-streams:3.5.1")
    implementation("org.apache.kafka:kafka-clients:3.5.1")

    // Jackson для сериализации/десериализации JSON
    implementation("com.fasterxml.jackson.core:jackson-databind:2.15.2")
    implementation("com.fasterxml.jackson.core:jackson-core:2.15.2")
    implementation("com.fasterxml.jackson.core:jackson-annotations:2.15.2")

    // SLF4J для логирования (рекомендуется для Kafka)
    implementation("org.slf4j:slf4j-api:2.0.9")
    implementation("org.slf4j:slf4j-simple:2.0.9")

    // Тестовые зависимости
    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testImplementation("org.apache.kafka:kafka-streams-test-utils:3.5.1")
}

tasks.test {
    useJUnitPlatform()
}

// Настройка Java версии
java {
    sourceCompatibility = JavaVersion.VERSION_17
    targetCompatibility = JavaVersion.VERSION_17
}