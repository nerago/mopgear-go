{
              "type": "TypeAPL",
              "simple": {
                "cooldowns": {
                  "hpPercentForDefensives": 0.3
                }
              },
              "prepullActions": [
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 20925
                      }
                    }
                  },
                  "doAtValue": {
                    "uuid": {
                      "value": "cab6766a-8a71-4ca3-9f69-1484a17c2447"
                    },
                    "const": {
                      "val": "-1.6s"
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "otherId": "OtherActionPotion"
                      }
                    }
                  },
                  "doAtValue": {
                    "uuid": {
                      "value": "5fe63640-0fa1-420c-aff7-b63e76ad89ec"
                    },
                    "const": {
                      "val": "-0.1s"
                    }
                  }
                }
              ],
              "priorityList": [
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 31884
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "9c2902a3-2e55-47d2-8b6b-6190101fdad1"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "76720847-7056-4fa4-8e10-4152e813dbc8"
                            },
                            "or": {
                              "vals": [
                                {
                                  "uuid": {
                                    "value": "b437b2ef-392c-47a4-b86e-5c48ea0a184b"
                                  },
                                  "and": {
                                    "vals": [
                                      {
                                        "uuid": {
                                          "value": "d012fe94-c510-4cb8-8972-b24902077471"
                                        },
                                        "bossCurrentTarget": {
                                          "targetUnit": {
                                            "type": "Target"
                                          }
                                        }
                                      },
                                      {
                                        "uuid": {
                                          "value": "ffffd300-1987-4685-b1f6-f68e6eb9acd8"
                                        },
                                        "cmp": {
                                          "op": "OpGt",
                                          "lhs": {
                                            "uuid": {
                                              "value": "cb761c9a-565b-4291-b5a6-3d0e6b222a9c"
                                            },
                                            "auraNumStacks": {
                                              "auraId": {
                                                "spellId": 144467
                                              },
                                              "includeReactionTime": true
                                            }
                                          },
                                          "rhs": {
                                            "uuid": {
                                              "value": "dc7376ad-6cfc-40be-be60-f64151da93fc"
                                            },
                                            "const": {
                                              "val": "3"
                                            }
                                          }
                                        }
                                      }
                                    ]
                                  }
                                },
                                {
                                  "uuid": {
                                    "value": "bac45be9-8c31-4c52-820e-8dc096be8615"
                                  },
                                  "cmp": {
                                    "op": "OpLe",
                                    "lhs": {
                                      "uuid": {
                                        "value": "dade7aea-8630-4841-9078-04f80f398b90"
                                      },
                                      "bossSpellTimeToReady": {
                                        "targetUnit": {
                                          "type": "Target"
                                        },
                                        "spellId": {
                                          "spellId": 144464
                                        }
                                      }
                                    },
                                    "rhs": {
                                      "uuid": {
                                        "value": "b708c896-634c-4eb9-8452-7bba0e63694f"
                                      },
                                      "const": {
                                        "val": "1s"
                                      }
                                    }
                                  }
                                }
                              ]
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 498
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "20fb2375-64ac-45b2-a3ca-87ef1762fe08"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "b1245056-6d42-4771-b05a-ab71188d3e20"
                            },
                            "not": {
                              "val": {
                                "uuid": {
                                  "value": "4d0cfe27-cc3b-44bc-9378-f5004838360d"
                                },
                                "spellIsReady": {
                                  "spellId": {
                                    "spellId": 498
                                  }
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "9062f003-2164-4e8c-8fae-ea21446c4a95"
                            },
                            "auraIsInactive": {
                              "auraId": {
                                "spellId": 498
                              },
                              "includeReactionTime": true
                            }
                          },
                          {
                            "uuid": {
                              "value": "21afda3d-b63c-4b74-8a00-8865b404197f"
                            },
                            "or": {
                              "vals": [
                                {
                                  "uuid": {
                                    "value": "e1a575ee-0103-4033-ac94-4805ad27f730"
                                  },
                                  "and": {
                                    "vals": [
                                      {
                                        "uuid": {
                                          "value": "bf5a56ba-8002-433c-ac9e-f805156619c8"
                                        },
                                        "bossCurrentTarget": {
                                          "targetUnit": {
                                            "type": "Target"
                                          }
                                        }
                                      },
                                      {
                                        "uuid": {
                                          "value": "70d741f8-3ce8-40b5-adf5-ffdfeab3db67"
                                        },
                                        "cmp": {
                                          "op": "OpGt",
                                          "lhs": {
                                            "uuid": {
                                              "value": "e5498e15-1d3c-49fa-b63e-db665c23f8b9"
                                            },
                                            "auraNumStacks": {
                                              "auraId": {
                                                "spellId": 144467
                                              },
                                              "includeReactionTime": true
                                            }
                                          },
                                          "rhs": {
                                            "uuid": {
                                              "value": "84bd8355-0043-4a58-878f-e10e5b8eeb81"
                                            },
                                            "const": {
                                              "val": "3"
                                            }
                                          }
                                        }
                                      }
                                    ]
                                  }
                                }
                              ]
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 86659
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "b2a0fe7a-2b11-412f-ae68-981c1487a7b3"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "ae5a5475-c2ba-4b82-a73b-ce2b31218ee3"
                            },
                            "bossCurrentTarget": {
                              "targetUnit": {
                                "type": "Target"
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "73ac9666-cb0b-483a-938f-cfb14a19392a"
                            },
                            "not": {
                              "val": {
                                "uuid": {
                                  "value": "113478aa-0b5f-4cc0-99f6-13fca3060235"
                                },
                                "spellIsReady": {
                                  "spellId": {
                                    "spellId": 86659
                                  }
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "562d3ed7-c905-44c3-96d3-b163ba225296"
                            },
                            "auraIsInactive": {
                              "auraId": {
                                "spellId": 86659
                              },
                              "includeReactionTime": true
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 31850
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "b2a39560-d644-47b8-b798-e77c3b825209"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "6b9f980b-b084-404c-9577-f961b109996d"
                            },
                            "bossCurrentTarget": {
                              "targetUnit": {
                                "type": "Target"
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "f45e2440-37c7-49ef-a7a3-c25331c96a57"
                            },
                            "not": {
                              "val": {
                                "uuid": {
                                  "value": "c897186e-df36-4c26-af83-61633c3e9416"
                                },
                                "spellIsReady": {
                                  "spellId": {
                                    "spellId": 31850
                                  }
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "bb90a06f-13c3-4456-b41f-c972184705b4"
                            },
                            "auraIsInactive": {
                              "auraId": {
                                "spellId": 31850
                              },
                              "includeReactionTime": true
                            }
                          }
                        ]
                      }
                    },
                    "castAllStatBuffCooldowns": {
                      "statType1": 11,
                      "statType2": 19,
                      "statType3": 10
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "7c318dba-0930-471e-9de5-b6dfff4dbfc8"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "b7353b06-b438-448d-ab0c-afb772bbbbb7"
                            },
                            "bossCurrentTarget": {
                              "targetUnit": {
                                "type": "Target"
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "14326b86-76d9-450a-9c77-2627086f72ae"
                            },
                            "not": {
                              "val": {
                                "uuid": {
                                  "value": "1ddab3c3-e81a-4e70-bd55-4cf44d4b7dd7"
                                },
                                "spellIsReady": {
                                  "spellId": {
                                    "spellId": 86659
                                  }
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "6ac4d004-3696-4b07-860a-ffc010c793e8"
                            },
                            "auraIsInactive": {
                              "auraId": {
                                "spellId": 86659
                              },
                              "includeReactionTime": true
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 114030,
                        "tag": -1
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "6b2f0ace-ff0b-4ebe-9184-156e1e3da979"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "caf6b59b-385f-41e8-90ae-97e80fa721eb"
                            },
                            "bossCurrentTarget": {
                              "targetUnit": {
                                "type": "Target"
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "263e1ed0-775a-49b1-8e54-8f8f4ba450ea"
                            },
                            "not": {
                              "val": {
                                "uuid": {
                                  "value": "463a6d87-57a5-4234-b621-246ea654d0fb"
                                },
                                "spellIsReady": {
                                  "spellId": {
                                    "spellId": 33206,
                                    "tag": -1
                                  }
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "f86bc515-6f06-47eb-a290-bdeeb9e9781f"
                            },
                            "auraIsInactive": {
                              "auraId": {
                                "spellId": 33206,
                                "tag": -1
                              },
                              "includeReactionTime": true
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 97462,
                        "tag": -1
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "de061025-cc9a-427c-a4aa-396c7a9e754c"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "d31685bc-78d1-4831-8147-592276af9511"
                            },
                            "bossCurrentTarget": {
                              "targetUnit": {
                                "type": "Target"
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "cf98f8ce-ac8c-455b-8a38-ec55db222333"
                            },
                            "cmp": {
                              "op": "OpLt",
                              "lhs": {
                                "uuid": {
                                  "value": "e0a74729-b295-429d-b592-f2e4c4ecdfc2"
                                },
                                "currentHealthPercent": {}
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "90687dfa-8830-4494-b8c1-bb20a6c2c78f"
                                },
                                "const": {
                                  "val": "50%"
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "itemId": 5512
                      }
                    }
                  }
                },
                {
                  "action": {
                    "autocastOtherCooldowns": {}
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 105809
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "9cdd38f6-efe7-4383-aa78-4c11c396dddc"
                      },
                      "or": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "494b218c-28a5-4c38-ab4a-f756d1480816"
                            },
                            "and": {
                              "vals": [
                                {
                                  "uuid": {
                                    "value": "79a668af-fb53-44c7-963a-6cb7777f31ea"
                                  },
                                  "cmp": {
                                    "op": "OpLt",
                                    "lhs": {
                                      "uuid": {
                                        "value": "413a01a7-fdad-48ff-8517-3bdf61707295"
                                      },
                                      "auraRemainingTime": {
                                        "auraId": {
                                          "spellId": 114163
                                        }
                                      }
                                    },
                                    "rhs": {
                                      "uuid": {
                                        "value": "e22578c2-7bd0-46af-8b5d-dcaa2466a850"
                                      },
                                      "const": {
                                        "val": "2s"
                                      }
                                    }
                                  }
                                },
                                {
                                  "uuid": {
                                    "value": "8815ab8b-0885-4fa7-a29c-0584f3e00177"
                                  },
                                  "cmp": {
                                    "op": "OpGt",
                                    "lhs": {
                                      "uuid": {
                                        "value": "e0c3bd86-5e6e-405e-9005-e9cf7d002bba"
                                      },
                                      "auraNumStacks": {
                                        "auraId": {
                                          "spellId": 114637
                                        }
                                      }
                                    },
                                    "rhs": {
                                      "uuid": {
                                        "value": "05cf0cfc-daf7-4f08-95a0-ed6072679f9b"
                                      },
                                      "const": {
                                        "val": "2"
                                      }
                                    }
                                  }
                                },
                                {
                                  "uuid": {
                                    "value": "d8408c1f-c057-49fa-be78-60e351e980c0"
                                  },
                                  "or": {
                                    "vals": [
                                      {
                                        "uuid": {
                                          "value": "13fdde5d-d485-4503-b54f-f25402b5d213"
                                        },
                                        "cmp": {
                                          "op": "OpGe",
                                          "lhs": {
                                            "uuid": {
                                              "value": "cb0ad1f8-7be1-4ed0-88d2-a794cdc05a95"
                                            },
                                            "currentGenericResource": {}
                                          },
                                          "rhs": {
                                            "uuid": {
                                              "value": "fa207116-521b-4075-a763-e9f1141a29e8"
                                            },
                                            "const": {
                                              "val": "3"
                                            }
                                          }
                                        }
                                      },
                                      {
                                        "uuid": {
                                          "value": "c1d886d7-5918-4cca-94e7-a9a69472fc3a"
                                        },
                                        "and": {
                                          "vals": [
                                            {
                                              "uuid": {
                                                "value": "e63f2db7-add4-422f-bc5d-88721bf785f7"
                                              },
                                              "auraIsKnown": {
                                                "auraId": {
                                                  "spellId": 144566
                                                }
                                              }
                                            },
                                            {
                                              "uuid": {
                                                "value": "0c53561f-f8e1-43b0-9590-29eb7517f59e"
                                              },
                                              "auraIsActive": {
                                                "auraId": {
                                                  "spellId": 144569
                                                },
                                                "includeReactionTime": true
                                              }
                                            }
                                          ]
                                        }
                                      }
                                    ]
                                  }
                                }
                              ]
                            }
                          },
                          {
                            "uuid": {
                              "value": "9198fd14-067d-4e67-9e3e-dc41ccf18347"
                            },
                            "and": {
                              "vals": [
                                {
                                  "uuid": {
                                    "value": "0d56f08a-9f54-4ecc-88c2-017938caa402"
                                  },
                                  "auraIsKnown": {
                                    "auraId": {
                                      "spellId": 144566
                                    }
                                  }
                                },
                                {
                                  "uuid": {
                                    "value": "5f83695d-2cd6-4b4c-a326-37dba8dc264f"
                                  },
                                  "auraIsActive": {
                                    "auraId": {
                                      "spellId": 144569
                                    },
                                    "includeReactionTime": true
                                  }
                                },
                                {
                                  "uuid": {
                                    "value": "0644df21-17e0-4990-81b8-ff0270a5202e"
                                  },
                                  "cmp": {
                                    "op": "OpGe",
                                    "lhs": {
                                      "uuid": {
                                        "value": "11f2335c-ef12-4aac-96ce-bcd8b906aa44"
                                      },
                                      "auraNumStacks": {
                                        "auraId": {
                                          "spellId": 114637
                                        }
                                      }
                                    },
                                    "rhs": {
                                      "uuid": {
                                        "value": "803f2f29-394b-4036-beea-c3a40e97ecfd"
                                      },
                                      "const": {
                                        "val": "5"
                                      }
                                    }
                                  }
                                }
                              ]
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 114163
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "f44d631c-e514-48e1-b970-0fe307125e01"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "d912f6cf-7e33-48dd-a41f-1929134a4d7a"
                            },
                            "cmp": {
                              "op": "OpGt",
                              "lhs": {
                                "uuid": {
                                  "value": "8b5f2746-4e33-4657-bd17-5b60ab5b0296"
                                },
                                "auraNumStacks": {
                                  "auraId": {
                                    "spellId": 114637
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "d700d60f-e854-4a76-abeb-80e377d13afc"
                                },
                                "const": {
                                  "val": "3"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "424251a4-b202-4b6c-bd7b-b05b08c109e6"
                            },
                            "cmp": {
                              "op": "OpLt",
                              "lhs": {
                                "uuid": {
                                  "value": "a4254529-ecb6-4353-8152-6396635d286f"
                                },
                                "currentHealthPercent": {}
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "6706d6b8-3463-457d-931d-adc3749d713f"
                                },
                                "const": {
                                  "val": "25"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "b92894b7-257d-4cc2-8716-8f3a3faf515b"
                            },
                            "cmp": {
                              "op": "OpGe",
                              "lhs": {
                                "uuid": {
                                  "value": "2d5ce783-97e5-4e80-90c9-f93edde581a2"
                                },
                                "currentGenericResource": {}
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "af84a7f4-022d-43b8-ad09-af8ae0239f79"
                                },
                                "const": {
                                  "val": "3"
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "castFriendlySpell": {
                      "spellId": {
                        "spellId": 85673
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "10cbf1f5-0b16-4613-8333-c6906f11fe16"
                      },
                      "or": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "802ecd6d-f10f-4774-9edc-8a412ba7c36b"
                            },
                            "cmp": {
                              "op": "OpLe",
                              "lhs": {
                                "uuid": {
                                  "value": "381af048-5fc6-4e41-bf89-d19fc92a6668"
                                },
                                "currentGenericResource": {}
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "c8370a3e-917d-417b-88ab-879a4f0ded16"
                                },
                                "const": {
                                  "val": "5"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "4b6f7fc4-9db7-4de3-b25c-af88e17dfa52"
                            },
                            "cmp": {
                              "op": "OpGe",
                              "lhs": {
                                "uuid": {
                                  "value": "4e700baa-80aa-43f8-9676-719da0fd3e30"
                                },
                                "protectionPaladinDamageTakenLastGlobal": {}
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "001bd82d-af55-4a80-b94c-e1ce2efdbbf9"
                                },
                                "math": {
                                  "op": "OpMul",
                                  "lhs": {
                                    "uuid": {
                                      "value": "d186fdfd-7eaa-45d5-9731-7f12053ca3f8"
                                    },
                                    "maxHealth": {}
                                  },
                                  "rhs": {
                                    "uuid": {
                                      "value": "2481a731-3b04-4faa-9eb8-4eb836945c6f"
                                    },
                                    "const": {
                                      "val": "0.3"
                                    }
                                  }
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 53600
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "fc2eb147-2698-43ba-83d6-2af2915a4fa6"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "4820d24f-7b6e-4b2d-a0fa-43f01e08ce48"
                            },
                            "auraIsKnown": {
                              "auraId": {
                                "spellId": 114232
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "a6b2026e-3dfd-4921-8768-2b5329282ebd"
                            },
                            "auraIsActive": {
                              "auraId": {
                                "spellId": 31884
                              }
                            }
                          }
                        ]
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 20271
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "ab9f4f2e-821a-4e7c-8838-37a09ede5c64"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "afd5c364-03c8-440c-a852-6c54343fcef6"
                            },
                            "auraIsKnown": {
                              "auraId": {
                                "spellId": 114232
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "a67f4529-672d-4782-83e0-a9301e9eddc8"
                            },
                            "cmp": {
                              "op": "OpGt",
                              "lhs": {
                                "uuid": {
                                  "value": "33822c25-363d-45aa-87bf-38098fa867ae"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 20271
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "4623a2bb-1a1c-42ca-bc6f-422e2c8af691"
                                },
                                "const": {
                                  "val": "0"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "87ee90c5-65b7-432e-bc2d-4235b1ed8a1a"
                            },
                            "cmp": {
                              "op": "OpLe",
                              "lhs": {
                                "uuid": {
                                  "value": "be9f979f-94d6-43cb-b25a-adb3316c713b"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 20271
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "2defc259-dacf-403d-aa09-cb05a15a7ce9"
                                },
                                "const": {
                                  "val": "0.5s"
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "wait": {
                      "duration": {
                        "uuid": {
                          "value": "bebd5a8b-a26a-47f2-b1c7-a7a32f4b0cb5"
                        },
                        "spellTimeToReady": {
                          "spellId": {
                            "spellId": 20271
                          }
                        }
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "95823b93-83e4-4b3d-b028-c7b3a50d3248"
                      },
                      "auraIsActive": {
                        "auraId": {
                          "spellId": 85416
                        },
                        "includeReactionTime": true
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 31935
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 20271
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 31935
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 119072
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "e0c9ab11-0b6c-4680-8c33-69e316f44d1e"
                      },
                      "auraIsActive": {
                        "auraId": {
                          "spellId": 105809
                        }
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 35395
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 114916
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 35395
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 24275
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "c741917a-ffeb-4249-8382-9ac95ae57538"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "c63ac0fb-f7a5-4f2c-b510-f503f90955c7"
                            },
                            "cmp": {
                              "op": "OpGt",
                              "lhs": {
                                "uuid": {
                                  "value": "770f59d5-d53f-41b8-9f1a-6e55ff42c457"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 35395
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "9b09ee12-552f-460e-9412-0a52ab9b916d"
                                },
                                "const": {
                                  "val": "0"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "44062780-9600-4835-aa13-f0c8a6d3f5a9"
                            },
                            "cmp": {
                              "op": "OpLe",
                              "lhs": {
                                "uuid": {
                                  "value": "79af3036-24c6-484e-8917-c81cf238d790"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 35395
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "aa338e8d-3a69-44d9-ba6d-8a776b61350b"
                                },
                                "const": {
                                  "val": "0.5s"
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "wait": {
                      "duration": {
                        "uuid": {
                          "value": "095636e3-4698-427f-8b6d-1227f7bb1c77"
                        },
                        "spellTimeToReady": {
                          "spellId": {
                            "spellId": 35395
                          }
                        }
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "2e888ec5-4182-493a-b6e0-735185095ef9"
                      },
                      "and": {
                        "vals": [
                          {
                            "uuid": {
                              "value": "b8ffdd04-9e1d-4df0-b755-b3248ea65b7a"
                            },
                            "cmp": {
                              "op": "OpGt",
                              "lhs": {
                                "uuid": {
                                  "value": "83c6b51c-11b8-44d1-b170-4670edda5520"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 20271
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "d702baf3-39c5-4ba0-9626-4259e4fb378c"
                                },
                                "const": {
                                  "val": "0"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "451202d8-1041-4000-8246-95008cd9b287"
                            },
                            "cmp": {
                              "op": "OpLe",
                              "lhs": {
                                "uuid": {
                                  "value": "f5a77368-b5ea-40e3-b383-d3404f2e76e7"
                                },
                                "spellTimeToReady": {
                                  "spellId": {
                                    "spellId": 20271
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "4bf866e6-26d7-428b-b173-c62ad53d84ec"
                                },
                                "const": {
                                  "val": "0.5s"
                                }
                              }
                            }
                          },
                          {
                            "uuid": {
                              "value": "1ab2af49-3019-43fd-93ae-77e8bb6b1bc2"
                            },
                            "cmp": {
                              "op": "OpGe",
                              "lhs": {
                                "uuid": {
                                  "value": "7f2f82b3-66c6-4f9c-b132-4f35aaa3500a"
                                },
                                "math": {
                                  "op": "OpSub",
                                  "lhs": {
                                    "uuid": {
                                      "value": "56b7c496-11a0-4e7d-af1a-19f5e4f62a31"
                                    },
                                    "spellTimeToReady": {
                                      "spellId": {
                                        "spellId": 35395
                                      }
                                    }
                                  },
                                  "rhs": {
                                    "uuid": {
                                      "value": "de8388dc-1863-4f3c-ae09-a29e4443beb8"
                                    },
                                    "spellTimeToReady": {
                                      "spellId": {
                                        "spellId": 20271
                                      }
                                    }
                                  }
                                }
                              },
                              "rhs": {
                                "uuid": {
                                  "value": "f51070cd-b06a-4109-9d85-7659d9eebd72"
                                },
                                "const": {
                                  "val": "0.5s"
                                }
                              }
                            }
                          }
                        ]
                      }
                    },
                    "wait": {
                      "duration": {
                        "uuid": {
                          "value": "9a90a763-770c-40c8-ac7c-d53149f6ef09"
                        },
                        "spellTimeToReady": {
                          "spellId": {
                            "spellId": 20271
                          }
                        }
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "15d52018-1706-45aa-9a94-04664e51fad7"
                      },
                      "cmp": {
                        "op": "OpLt",
                        "lhs": {
                          "uuid": {
                            "value": "f6c26412-4236-4c0b-9032-5ec090dcc34e"
                          },
                          "auraRemainingTime": {
                            "auraId": {
                              "spellId": 20925
                            }
                          }
                        },
                        "rhs": {
                          "uuid": {
                            "value": "99e11da1-b135-4f55-9484-599c4bcdf28e"
                          },
                          "const": {
                            "val": "5s"
                          }
                        }
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 20925
                      }
                    }
                  }
                },
                {
                  "action": {
                    "condition": {
                      "uuid": {
                        "value": "d7ca8905-b814-4852-b15f-29e5d1f8b338"
                      },
                      "not": {
                        "val": {
                          "uuid": {
                            "value": "59661502-847e-46e7-8811-b7fd63f25507"
                          },
                          "dotIsActive": {
                            "spellId": {
                              "spellId": 26573
                            }
                          }
                        }
                      }
                    },
                    "castSpell": {
                      "spellId": {
                        "spellId": 26573
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 114158
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 114852,
                        "tag": 1
                      }
                    }
                  }
                },
                {
                  "action": {
                    "castSpell": {
                      "spellId": {
                        "spellId": 20925
                      }
                    }
                  }
                }
              ]
            }