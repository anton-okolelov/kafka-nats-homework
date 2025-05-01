package com.okolelov;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategy;
import org.apache.kafka.common.errors.SerializationException;
import org.apache.kafka.common.serialization.Deserializer;
import org.apache.kafka.common.serialization.Serde;
import org.apache.kafka.common.serialization.Serdes;
import org.apache.kafka.common.serialization.Serializer;
import org.apache.kafka.common.utils.Bytes;
import org.apache.kafka.streams.KafkaStreams;
import org.apache.kafka.streams.KeyValue;
import org.apache.kafka.streams.StreamsBuilder;
import org.apache.kafka.streams.StreamsConfig;
import org.apache.kafka.streams.kstream.*;
import org.apache.kafka.streams.state.WindowStore;

import java.time.Duration;
import java.util.Properties;

public class ClickThroughRateProcessor {

    public static void main(String[] args) {
        Properties props = new Properties();
        props.put(StreamsConfig.APPLICATION_ID_CONFIG, "ctr-calcul");
        props.put(StreamsConfig.BOOTSTRAP_SERVERS_CONFIG, "localhost:10002,localhost:10003,localhost:10004");

        // Указываем, что мы хотим использовать JSON сериализацию
        props.put(StreamsConfig.DEFAULT_KEY_SERDE_CLASS_CONFIG, Serdes.String().getClass());
        props.put(StreamsConfig.DEFAULT_VALUE_SERDE_CLASS_CONFIG, Serdes.String().getClass());

        // Для обеспечения точности расчетов в скользящем окне
        props.put(StreamsConfig.COMMIT_INTERVAL_MS_CONFIG, 1000);

        StreamsBuilder builder = new StreamsBuilder();

        // Регистрируем сериализаторы для наших типов
        final Serde<ClickEvent> clickEventSerde = new ClickEventSerde();
        final Serde<ViewEvent> viewEventSerde = new ViewEventSerde();
        final Serde<CTRResult> ctrResultSerde = new CTRResultSerde();

        // Читаем данные из топиков
        KStream<String, ClickEvent> clicksStream = builder.stream(
                "clicks_v1",
                Consumed.with(Serdes.String(), clickEventSerde)
        );

        KStream<String, ViewEvent> viewsStream = builder.stream(
                "views_v1",
                Consumed.with(Serdes.String(), viewEventSerde)
        );

        // Преобразуем потоки, чтобы использовать advertisingId в качестве ключа
        KStream<String, ClickEvent> clicksKeyedStream = clicksStream
                .selectKey((key, value) -> value.getAdvertisingId());

        KStream<String, ViewEvent> viewsKeyedStream = viewsStream
                .selectKey((key, value) -> value.getAdvertisingId());

        // Создаем временные окна в 10 минут для подсчета
        TimeWindows timeWindows = TimeWindows.of(Duration.ofSeconds(20));

        // Группируем и считаем клики по advertisingId в окнах
        KTable<Windowed<String>, Long> clickCounts = clicksKeyedStream
                .groupByKey()
                .windowedBy(timeWindows)
                .count(Materialized.<String, Long, WindowStore<Bytes, byte[]>>as("click-counts")
                        .withKeySerde(Serdes.String())
                        .withValueSerde(Serdes.Long()));

        // Группируем и считаем просмотры по advertisingId в окнах
        KTable<Windowed<String>, Long> viewCounts = viewsKeyedStream
                .groupByKey()
                .windowedBy(timeWindows)
                .count(Materialized.<String, Long, WindowStore<Bytes, byte[]>>as("view-counts")
                        .withKeySerde(Serdes.String())
                        .withValueSerde(Serdes.Long()));

        // Объединяем результаты и вычисляем CTR
        KTable<Windowed<String>, CTRResult> ctrTable = clickCounts
                .join(
                        viewCounts,
                        (clickCount, viewCount) -> {
                            // Здесь мы не можем получить ключ окна напрямую,
                            // поэтому мы создаем CTRResult с временными значениями,
                            // которые будут заменены в следующем шаге
                            return new CTRResult("temp", clickCount, viewCount,
                                    viewCount > 0 ? (double) clickCount / viewCount : 0.0, 0L, 0L);
                        }
                );

        // Преобразуем обратно в поток и заполняем информацию о ключе и окне
        ctrTable.toStream()
                .map((windowed, value) -> {
                    // Здесь мы можем получить доступ к информации об окне
                    String advertisingId = windowed.key();
                    long windowStart = windowed.window().start();
                    long windowEnd = windowed.window().end();

                    // Создаем новый CTRResult с правильными данными
                    CTRResult updatedResult = new CTRResult(
                            advertisingId,
                            value.getClicks(),
                            value.getViews(),
                            value.getCtr(),
                            windowStart,
                            windowEnd
                    );

                    return KeyValue.pair(advertisingId, updatedResult);
                })
                .peek((key, value) -> {
                    System.out.println("Sending to ctr_v1 topic: " +
                            "Key=" + key + ", " +
                            "Value=" + value);
                })
                .to("ctr_v1", Produced.with(Serdes.String(), ctrResultSerde));

        // Строим и запускаем приложение Kafka Streams
        KafkaStreams streams = new KafkaStreams(builder.build(), props);
        streams.start();

        // Добавляем обработчик для корректного завершения
        Runtime.getRuntime().addShutdownHook(new Thread(streams::close));
    }

    // Определение классов для представления событий
    public static class ClickEvent {
        @JsonProperty("advertising_id")
        private String advertisingId;
        @JsonProperty("time_stamp")
        private long timestamp;

        public ClickEvent() {
            // Пустой конструктор для Jackson
        }

        public ClickEvent(String advertisingId, long timestamp) {
            this.advertisingId = advertisingId;
            this.timestamp = timestamp;
        }

        public String getAdvertisingId() {
            return advertisingId;
        }

        public void setAdvertisingId(String advertisingId) {
            this.advertisingId = advertisingId;
        }

        public long getTimestamp() {
            return timestamp;
        }

        public void setTimestamp(long timestamp) {
            this.timestamp = timestamp;
        }
    }

    public static class ViewEvent {
        @JsonProperty("advertising_id")
        private String advertisingId;
        @JsonProperty("time_stamp")
        private long timestamp;

        public ViewEvent() {
            // Пустой конструктор для Jackson
        }

        public ViewEvent(String advertisingId, long timestamp) {
            this.advertisingId = advertisingId;
            this.timestamp = timestamp;
        }

        public String getAdvertisingId() {
            return advertisingId;
        }

        public void setAdvertisingId(String advertisingId) {
            this.advertisingId = advertisingId;
        }

        public long getTimestamp() {
            return timestamp;
        }

        public void setTimestamp(long timestamp) {
            this.timestamp = timestamp;
        }
    }

    public static class CTRResult {
        private String advertisingId;
        private long clicks;
        private long views;
        private double ctr;
        private long windowStart;
        private long windowEnd;

        public CTRResult() {
            // Пустой конструктор для Jackson
        }

        public CTRResult(String advertisingId, long clicks, long views, double ctr, long windowStart, long windowEnd) {
            this.advertisingId = advertisingId;
            this.clicks = clicks;
            this.views = views;
            this.ctr = ctr;
            this.windowStart = windowStart;
            this.windowEnd = windowEnd;
        }

        public String getAdvertisingId() {
            return advertisingId;
        }

        public void setAdvertisingId(String advertisingId) {
            this.advertisingId = advertisingId;
        }

        public long getClicks() {
            return clicks;
        }

        public void setClicks(long clicks) {
            this.clicks = clicks;
        }

        public long getViews() {
            return views;
        }

        public void setViews(long views) {
            this.views = views;
        }

        public double getCtr() {
            return ctr;
        }

        public void setCtr(double ctr) {
            this.ctr = ctr;
        }

        public long getWindowStart() {
            return windowStart;
        }

        public void setWindowStart(long windowStart) {
            this.windowStart = windowStart;
        }

        public long getWindowEnd() {
            return windowEnd;
        }

        public void setWindowEnd(long windowEnd) {
            this.windowEnd = windowEnd;
        }

        @Override
        public String toString() {
            return "CTRResult{" +
                    "advertisingId='" + advertisingId + '\'' +
                    ", clicks=" + clicks +
                    ", views=" + views +
                    ", ctr=" + ctr +
                    ", windowStart=" + windowStart +
                    ", windowEnd=" + windowEnd +
                    '}';
        }
    }

    // Сериализаторы и десериализаторы для наших объектов
    public static class ClickEventSerde extends WrapperSerde<ClickEvent> {
        public ClickEventSerde() {
            super(new JsonSerializer<>(), new JsonDeserializer<>(ClickEvent.class));
        }
    }

    public static class ViewEventSerde extends WrapperSerde<ViewEvent> {
        public ViewEventSerde() {
            super(new JsonSerializer<>(), new JsonDeserializer<>(ViewEvent.class));
        }
    }

    public static class CTRResultSerde extends WrapperSerde<CTRResult> {
        public CTRResultSerde() {
            super(new JsonSerializer<>(), new JsonDeserializer<>(CTRResult.class));
        }
    }

    // Абстрактный класс для сериализации/десериализации
    private static class WrapperSerde<T> implements Serde<T> {
        private JsonSerializer<T> serializer;
        private JsonDeserializer<T> deserializer;

        WrapperSerde(JsonSerializer<T> serializer, JsonDeserializer<T> deserializer) {
            this.serializer = serializer;
            this.deserializer = deserializer;
        }

        @Override
        public Serializer<T> serializer() {
            return serializer;
        }

        @Override
        public Deserializer<T> deserializer() {
            return deserializer;
        }
    }

    // Для простоты примера, реализуем базовый JSON сериализатор/десериализатор
    private static class JsonSerializer<T> implements Serializer<T> {
        private final ObjectMapper mapper;

        public JsonSerializer() {
            this.mapper = new ObjectMapper();
            // Настраиваем ObjectMapper для поддержки snake_case
            mapper.setPropertyNamingStrategy(PropertyNamingStrategy.SNAKE_CASE);
        }

        @Override
        public byte[] serialize(String topic, T data) {
            if (data == null) return null;
            try {
                return mapper.writeValueAsBytes(data);
            } catch (Exception e) {
                throw new SerializationException("Error serializing JSON", e);
            }
        }
    }

    private static class JsonDeserializer<T> implements Deserializer<T> {
        private final ObjectMapper mapper;
        private final Class<T> tClass;

        public JsonDeserializer(Class<T> tClass) {
            this.tClass = tClass;
            this.mapper = new ObjectMapper();
            // Настраиваем ObjectMapper для поддержки snake_case
            mapper.setPropertyNamingStrategy(PropertyNamingStrategy.SNAKE_CASE);
        }

        @Override
        public T deserialize(String topic, byte[] data) {
            if (data == null) return null;
            try {
                return mapper.readValue(data, tClass);
            } catch (Exception e) {
                throw new SerializationException("Error deserializing JSON", e);
            }
        }
    }
}