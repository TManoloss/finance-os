import asyncio
import unittest

from app.models.transaction import TransactionClassifyRequest
from app.services.classifier import ClassifierService


class ClassifierHierarchyTest(unittest.TestCase):
    def test_specific_subcategory_beats_generic_transfer_and_legacy_rules(self):
        cases = (
            ("PIX ENVIADO | UBER", "Transporte por aplicativo"),
            ("PIX ENVIADO | SHOPEE", "Marketplace"),
        )

        for description, expected in cases:
            with self.subTest(description=description):
                result = asyncio.run(
                    ClassifierService().classify(
                        TransactionClassifyRequest(
                            merchant_name="",
                            description=description,
                            amount=42,
                            direction="debit",
                        )
                    )
                )
                self.assertEqual(result["category_id"], expected)
                self.assertEqual(result["confidence"], 1.0)


if __name__ == "__main__":
    unittest.main()
