#!/usr/bin/env python3
import os
import re
import csv
import json
import struct
import math
import random

MAGIC_GOW = b'GOW\x01'

def tokenize(text):
    words = [w.lower() for w in re.findall(r'[\w]+', text) if w]
    tokens = list(words)
    for w in words:
        if len(w) >= 2:
            for i in range(len(w) - 1):
                tokens.append(w[i:i+2])
    return tokens

def transpile_ngo(src):
    train_pattern = re.compile(r'train\s+"([^"]+)"\s*\{([^}]+)\}', re.DOTALL)
    train_configs = []
    
    def repl_train(m):
        weight_path = m.group(1)
        body = m.group(2)
        cfg = {"weight_path": weight_path, "epochs": 50, "format": "csv", "input": "text", "target": "label"}
        for line in body.strip().split("\n"):
            line = line.strip()
            if line.startswith("format:"):
                cfg["format"] = line.split(":", 1)[1].strip(' ",;')
            elif line.startswith("source:"):
                cfg["source"] = line.split(":", 1)[1].strip(' ",;')
            elif line.startswith("input:"):
                cfg["input"] = line.split(":", 1)[1].strip(' ",;')
            elif line.startswith("target:"):
                cfg["target"] = line.split(":", 1)[1].strip(' ",;')
            elif line.startswith("epochs:"):
                cfg["epochs"] = int(re.findall(r'\d+', line)[0])
        train_configs.append(cfg)
        return f'/* [NeuroGo] train "{weight_path}" extracted */'
    
    src_no_train = train_pattern.sub(repl_train, src)
    
    match_pattern = re.compile(r'match\s+(.+?)\s+using\s+"([^"]+)"\s*\{', re.DOTALL)
    case_pattern = re.compile(r'case\s+"([^"]+)"\s+score\s*(>=|>|<=|<|==)\s*([0-9.]+)\s*:', re.DOTALL)
    
    out = src_no_train
    while True:
        m = match_pattern.search(out)
        if not m:
            break
        h_start, h_end = m.span()
        expr = m.group(1).strip()
        weight_path = m.group(2).strip()
        
        brace = 1
        b_end = -1
        for i in range(h_end, len(out)):
            if out[i] == '{':
                brace += 1
            elif out[i] == '}':
                brace -= 1
                if brace == 0:
                    b_end = i
                    break
        if b_end == -1:
            raise ValueError("Unmatched brace in match block")
        
        body = out[h_end:b_end]
        def repl_case(cm):
            lbl = cm.group(1)
            op = cm.group(2)
            thresh = cm.group(3)
            return f'case _ngoMatch.Label == "{lbl}" && _ngoMatch.Score {op} {thresh}:'
        
        new_body = case_pattern.sub(repl_case, body)
        replacement = f'{{\n\t_ngoMatch := runtime.Match("{weight_path}", {expr})\n\tswitch {{\n{new_body}\n\t}}\n}}'
        out = out[:h_start] + replacement + out[b_end+1:]
        
    if '"neurogo/pkg/runtime"' not in out:
        pkg_pos = out.find("package ")
        if pkg_pos != -1:
            nl_pos = out.find("\n", pkg_pos)
            out = out[:nl_pos+1] + '\nimport "neurogo/pkg/runtime"\n' + out[nl_pos+1:]
            
    return out, train_configs

def train_model(cfg, base_dir="."):
    csv_path = os.path.join(base_dir, cfg["source"])
    weight_path = os.path.join(base_dir, cfg["weight_path"])
    
    print(f"[Trainer] Training from '{csv_path}' -> '{weight_path}'")
    samples = []
    with open(csv_path, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            samples.append((row[cfg["input"]].strip('"'), row[cfg["target"]].strip('"')))
            
    labels = sorted(list(set(s[1] for s in samples)))
    label_to_id = {l: i for i, l in enumerate(labels)}
    
    vocab = {}
    for text, _ in samples:
        for tok in tokenize(text):
            if tok not in vocab:
                vocab[tok] = len(vocab)
                
    num_classes = len(labels)
    input_dim = len(vocab)
    print(f"[Trainer] Samples: {len(samples)}, Classes: {labels}, Vocab size: {input_dim}")
    
    W = [[random.uniform(-0.05, 0.05) for _ in range(num_classes)] for _ in range(input_dim)]
    B = [0.0 for _ in range(num_classes)]
    
    lr = 0.2
    epochs = cfg.get("epochs", 50)
    for ep in range(1, epochs + 1):
        random.shuffle(samples)
        total_loss = 0.0
        for text, lbl in samples:
            target = label_to_id[lbl]
            tokens = tokenize(text)
            
            logits = list(B)
            for tok in tokens:
                idx = vocab[tok]
                for c in range(num_classes):
                    logits[c] += W[idx][c]
            
            max_l = max(logits)
            exp_l = [math.exp(x - max_l) for x in logits]
            sum_exp = sum(exp_l)
            probs = [x / sum_exp for x in exp_l]
            
            loss = -math.log(max(probs[target], 1e-12))
            total_loss += loss
            
            for c in range(num_classes):
                grad = probs[c] - (1.0 if c == target else 0.0)
                B[c] -= lr * grad
                for tok in tokens:
                    idx = vocab[tok]
                    W[idx][c] -= lr * grad
                    
        if ep % 10 == 0 or ep == epochs:
            print(f"   Epoch [{ep}/{epochs}] - Loss: {total_loss/len(samples):.4f}")
            
    header = {
        "magic": [71, 79, 87, 1],
        "version": 1,
        "architecture": "linear_softmax",
        "input_dim": input_dim,
        "num_classes": num_classes,
        "labels": labels,
        "vocab": vocab
    }
    header_json = json.dumps(header).encode('utf-8')
    
    with open(weight_path, 'wb') as f:
        f.write(MAGIC_GOW)
        f.write(struct.pack('<I', len(header_json)))
        f.write(header_json)
        flat_w = [W[i][c] for i in range(input_dim) for c in range(num_classes)]
        f.write(struct.pack(f'<{len(flat_w)}f', *flat_w))
        f.write(struct.pack(f'<{len(B)}f', *B))
        
    print(f"[Trainer] Model successfully saved to '{weight_path}' ({os.path.getsize(weight_path)} bytes)")

def load_and_infer(weight_path, text):
    with open(weight_path, 'rb') as f:
        magic = f.read(4)
        if magic != MAGIC_GOW:
            raise ValueError("Invalid magic in .gow")
        header_len = struct.unpack('<I', f.read(4))[0]
        header_json = f.read(header_len).decode('utf-8')
        header = json.loads(header_json)
        
        input_dim = header["input_dim"]
        num_classes = header["num_classes"]
        labels = header["labels"]
        vocab = header["vocab"]
        
        flat_w = struct.unpack(f'<{input_dim * num_classes}f', f.read(input_dim * num_classes * 4))
        biases = struct.unpack(f'<{num_classes}f', f.read(num_classes * 4))
        
    tokens = tokenize(text)
    logits = list(biases)
    for tok in tokens:
        if tok in vocab:
            idx = vocab[tok]
            for c in range(num_classes):
                logits[c] += flat_w[idx * num_classes + c]
                
    max_l = max(logits)
    exp_l = [math.exp(x - max_l) for x in logits]
    sum_exp = sum(exp_l)
    probs = [x / sum_exp for x in exp_l]
    
    best_idx = max(range(num_classes), key=lambda i: probs[i])
    return {
        "label": labels[best_idx],
        "score": probs[best_idx],
        "scores": {labels[i]: probs[i] for i in range(num_classes)}
    }

if __name__ == "__main__":
    example_dir = "/working_dir/c_b5c449c5c377d857/neurogo/examples/intent"
    ngo_path = os.path.join(example_dir, "main.ngo")
    go_path = os.path.join(example_dir, "main.go")
    
    print("=== Step 1: Transpiling main.ngo -> main.go ===")
    with open(ngo_path, 'r', encoding='utf-8') as f:
        src = f.read()
    go_code, configs = transpile_ngo(src)
    with open(go_path, 'w', encoding='utf-8') as f:
        f.write(go_code)
    print(f"Transpiled code written to {go_path}")
    
    print("\n=== Step 2: Training Model defined in main.ngo ===")
    for cfg in configs:
        train_model(cfg, base_dir=example_dir)
        
    print("\n=== Step 3: Testing Intelligent Match & Branching ===")
    weight_file = os.path.join(example_dir, configs[0]["weight_path"])
    
    test_queries = [
        "주문 결제한 거 환불해주세요",
        "송장번호 배송 조회 좀 해주세요",
        "이거 사이즈 재고 있나요?",
        "오늘 점심 뭐 먹지?"
    ]
    
    for q in test_queries:
        res = load_and_infer(weight_file, q)
        print(f"\n[User Query] \"{q}\"")
        print(f"   -> Top Prediction: {res['label']} (Score: {res['score']:.4f})")
        print(f"   -> Probabilities: { {k: round(v, 4) for k, v in res['scores'].items()} }")
        
        score = res['score']
        label = res['label']
        if label == "Refund" and score >= 0.70:
            print("   >> [Branch Taken] case \"Refund\" score >= 0.70: 환불/결제취소 전문 상담사로 연결합니다.")
        elif label == "Delivery" and score >= 0.70:
            print("   >> [Branch Taken] case \"Delivery\" score >= 0.70: 실시간 배송 추적 시스템을 실행합니다.")
        elif label == "Inquiry" and score >= 0.70:
            print("   >> [Branch Taken] case \"Inquiry\" score >= 0.70: 상품 상세 FAQ 및 Q&A 봇으로 안내합니다.")
        else:
            print("   >> [Branch Taken] default: 명확하지 않은 문의입니다. 기본 고객센터로 연결합니다.")
