#!/usr/bin/env python3
"""
ArticleToUse - Paper Relevance Scorer

Calcula score de relevância de papers do arXiv para o BusinessOS.
"""

import json
import re
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass


@dataclass
class RelevanceScore:
    """Breakdown do score de relevância"""
    technology_match: float  # 0-25
    feature_alignment: float  # 0-30
    implementation_feasibility: float  # 0-20
    innovation_potential: float  # 0-15
    recency_relevance: float  # 0-10

    @property
    def total(self) -> float:
        return (
            self.technology_match +
            self.feature_alignment +
            self.implementation_feasibility +
            self.innovation_potential +
            self.recency_relevance
        )


class PaperScorer:
    """Calcula relevância de papers para o BusinessOS"""

    def __init__(self, config_path: str = "config/businessos_features.json"):
        self.config_path = Path(config_path)
        self.config = self._load_config()

    def _load_config(self) -> Dict:
        """Carrega configuração de features do BusinessOS"""
        with open(self.config_path) as f:
            return json.load(f)

    def score_paper(self, title: str, abstract: str, published_date: str) -> RelevanceScore:
        """Calcula score total de um paper"""
        text = f"{title.lower()} {abstract.lower()}"
        pub_date = datetime.fromisoformat(published_date.replace('Z', '+00:00'))

        return RelevanceScore(
            technology_match=self._score_technology_match(text),
            feature_alignment=self._score_feature_alignment(text),
            implementation_feasibility=self._score_feasibility(text, abstract),
            innovation_potential=self._score_innovation(abstract),
            recency_relevance=self._score_recency(pub_date)
        )

    def _score_technology_match(self, text: str) -> float:
        """Score baseado em match de tecnologias (0-25)"""
        score = 0.0

        # Tier 1: 5 pontos cada
        tier1 = self.config["technologies"]["tier1"]["keywords"]
        for keyword in tier1:
            if keyword.lower() in text:
                score += 5.0

        # Tier 2: 3 pontos cada
        tier2 = self.config["technologies"]["tier2"]["keywords"]
        for keyword in tier2:
            if keyword.lower() in text:
                score += 3.0

        # Tier 3: 1 ponto cada
        tier3 = self.config["technologies"]["tier3"]["keywords"]
        for keyword in tier3:
            if keyword.lower() in text:
                score += 1.0

        return min(score, 25.0)

    def _score_feature_alignment(self, text: str) -> float:
        """Score baseado em alinhamento com features do BusinessOS (0-30)"""
        score = 0.0

        features = self.config["features"]
        for feature_key, feature_data in features.items():
            keywords = feature_data["keywords"]
            matches = sum(1 for kw in keywords if kw.lower() in text)

            if matches > 0:
                # 10 pontos máximo por feature, proporcional aos matches
                match_ratio = matches / len(keywords)
                score += 10.0 * match_ratio

        return min(score, 30.0)

    def _score_feasibility(self, text: str, abstract: str) -> float:
        """Score de viabilidade de implementação (0-20)"""
        score = 10.0  # baseline

        # Sinais positivos
        if any(word in text for word in ["implementation", "code", "github", "open source"]):
            score += 5.0

        if any(word in text for word in ["postgresql", "python", "go", "typescript"]):
            score += 3.0

        # Sinais negativos
        if any(word in text for word in ["theoretical", "complex architecture", "requires large dataset"]):
            score -= 3.0

        return max(0, min(score, 20.0))

    def _score_innovation(self, abstract: str) -> float:
        """Score de potencial de inovação (0-15)"""
        score = 5.0  # baseline

        # Indicadores de performance
        performance_keywords = [
            "outperforms", "state-of-the-art", "sota",
            "significant improvement", "breakthrough", "superior"
        ]
        for kw in performance_keywords:
            if kw in abstract.lower():
                score += 2.0

        # Indicadores de novidade
        novelty_keywords = ["novel", "new approach", "first", "introduce", "pioneer"]
        for kw in novelty_keywords:
            if kw in abstract.lower():
                score += 1.5

        return min(score, 15.0)

    def _score_recency(self, pub_date: datetime) -> float:
        """Score baseado em quão recente é o paper (0-10)"""
        now = datetime.now(pub_date.tzinfo)
        age = now - pub_date

        if age < timedelta(days=30):
            return 10.0
        elif age < timedelta(days=90):
            return 7.0
        elif age < timedelta(days=180):
            return 4.0
        elif age < timedelta(days=365):
            return 2.0
        else:
            return 0.0

    def generate_application_ideas(
        self,
        title: str,
        abstract: str,
        score: RelevanceScore
    ) -> List[Dict]:
        """Gera ideias de aplicação baseado no paper"""
        ideas = []
        text = f"{title.lower()} {abstract.lower()}"

        # Mapeamento de keywords para features
        feature_mappings = {
            "query expansion": {
                "feature": "rag_enhancement",
                "title": "Enhance Query Expansion Service",
                "files": ["desktop/backend-go/internal/services/query_expansion.go"],
                "complexity": "medium",
                "effort": "2-4 days"
            },
            "re-ranking": {
                "feature": "rag_enhancement",
                "title": "Improve Re-Ranking Algorithm",
                "files": ["desktop/backend-go/internal/services/reranker.go"],
                "complexity": "medium",
                "effort": "3-5 days"
            },
            "memory": {
                "feature": "memory_hierarchy",
                "title": "Enhance Memory Hierarchy",
                "files": ["desktop/backend-go/internal/services/memory_hierarchy_service.go"],
                "complexity": "high",
                "effort": "5-7 days"
            },
            "agent": {
                "feature": "agent_orchestration",
                "title": "Improve Agent Orchestration",
                "files": ["desktop/backend-go/internal/services/orchestrator.go"],
                "complexity": "high",
                "effort": "4-6 days"
            }
        }

        for keyword, mapping in feature_mappings.items():
            if keyword in text:
                ideas.append({
                    "feature_area": mapping["feature"],
                    "title": mapping["title"],
                    "description": f"Apply techniques from this paper to {mapping['feature']}",
                    "target_files": mapping["files"],
                    "implementation_difficulty": mapping["complexity"],
                    "estimated_effort": mapping["effort"],
                    "priority": min(10, int(score.total / 10)),
                    "impact": "High" if score.total > 70 else "Medium"
                })

        return ideas


def main():
    """Exemplo de uso"""
    scorer = PaperScorer()

    # Exemplo de paper
    title = "Hierarchical Memory Systems for Large Language Models with Adaptive RAG"
    abstract = """
    We propose a novel hierarchical memory architecture for LLMs that combines
    semantic search with query expansion and re-ranking. Our approach outperforms
    state-of-the-art RAG systems by 30% on benchmark tasks. We demonstrate
    significant improvements in retrieval accuracy using pgvector for efficient
    vector search. Implementation available on GitHub.
    """
    published_date = "2026-01-05T00:00:00Z"

    score = scorer.score_paper(title, abstract, published_date)
    ideas = scorer.generate_application_ideas(title, abstract, score)

    print("=" * 70)
    print("RELEVANCE SCORE REPORT")
    print("=" * 70)
    print(f"\nTitle: {title}\n")
    print(f"Total Score: {score.total:.1f}/100\n")
    print("Breakdown:")
    print(f"  Technology Match:      {score.technology_match:.1f}/25")
    print(f"  Feature Alignment:     {score.feature_alignment:.1f}/30")
    print(f"  Feasibility:           {score.implementation_feasibility:.1f}/20")
    print(f"  Innovation:            {score.innovation_potential:.1f}/15")
    print(f"  Recency:               {score.recency_relevance:.1f}/10")

    if score.total >= 70:
        print(f"\n[HIGH] RELEVANCE - Implement ASAP!")
    elif score.total >= 50:
        print(f"\n[MEDIUM] RELEVANCE - Add to watch list")
    elif score.total >= 40:
        print(f"\n[LOW] RELEVANCE - Reference only")
    else:
        print(f"\n[FILTERED] BELOW THRESHOLD - Not relevant")

    if ideas:
        print(f"\n\nAPPLICATION IDEAS ({len(ideas)}):")
        for i, idea in enumerate(ideas, 1):
            print(f"\n{i}. {idea['title']}")
            print(f"   Feature: {idea['feature_area']}")
            print(f"   Complexity: {idea['implementation_difficulty']}")
            print(f"   Effort: {idea['estimated_effort']}")
            print(f"   Priority: {idea['priority']}/10")
            print(f"   Files: {', '.join(idea['target_files'])}")

    print("\n" + "=" * 70)


if __name__ == "__main__":
    main()
