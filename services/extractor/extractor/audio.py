"""Audio extraction — transcription via Whisper (SGLang or dedicated instance).

This module is the foundation for audio/video ingestion (Phase 10).
Whisper model will be served via a dedicated SGLang instance (sglang-whisper).
"""

import logging

import httpx

logger = logging.getLogger(__name__)


class AudioTranscriber:
    """Transcribes audio files via a Whisper-compatible API.

    When the Whisper SGLang instance is configured, this sends audio
    to the /v1/audio/transcriptions endpoint (OpenAI-compatible).
    """

    def __init__(self, endpoint: str, model: str = "openai/whisper-large-v3"):
        self.endpoint = endpoint.rstrip("/")
        self.model = model
        self._client = httpx.Client(timeout=300.0)  # audio can be long

    def transcribe(self, audio_bytes: bytes, language: str = "es") -> str:
        """Transcribe audio bytes to text.

        Args:
            audio_bytes: Raw audio file bytes (mp3, wav, etc.)
            language: Language hint for the model.

        Returns:
            Transcribed text.
        """
        resp = self._client.post(
            f"{self.endpoint}/v1/audio/transcriptions",
            files={"file": ("audio.mp3", audio_bytes, "audio/mpeg")},
            data={"model": self.model, "language": language},
        )
        resp.raise_for_status()
        return resp.json().get("text", "")

    def close(self):
        self._client.close()
