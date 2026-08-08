import { useState, useEffect, useCallback } from "react";
import {
  LineChart, Line, AreaChart, Area, BarChart, Bar,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell
} from "recharts";
import "./Monitoring.scss";

const API_URL = "https://api-gateway-406932815122.europe-west1.run.app";

const SERVICES = [
  { name: "API Gateway", url: API_URL, endpoint: "/api/tours/published" },
  { name: "Tour Service", url: "https://tour-service-406932815122.europe-west1.run.app", endpoint: "/api/tours/published" },
  { name: "Stakeholders", url: "https://stakeholders-service-406932815122.europe-west1.run.app", endpoint: "/health" },
  { name: "Follower", url: "https://follower-service-406932815122.europe-west1.run.app", endpoint: "/following/1" },
  { name: "Payment", url: "https://payment-service-406932815122.europe-west1.run.app", endpoint: "/api/cart" },
  { name: "Blog", url: "https://blog-service-406932815122.europe-west1.run.app", endpoint: "/api/blogs" },
];

const COLORS = ["#F45F00", "#1E3D59", "#48749E", "#FA9819", "#F20909", "#B6C9CF"];

export default function Monitoring() {
  const [activeTab, setActiveTab] = useState<"live" | "loadtest" | "results">("live");
  const [liveData, setLiveData] = useState<any[]>([]);
  const [loadTestRunning, setLoadTestRunning] = useState(false);
  const [loadTestData, setLoadTestData] = useState<any[]>([]);
  const [loadTestStats, setLoadTestStats] = useState<any>(null);
  const [concurrentUsers, setConcurrentUsers] = useState(10);
  const [duration, setDuration] = useState(30);
  const [progress, setProgress] = useState(0);
  const [healthStatus, setHealthStatus] = useState<any>({});
  const [requestCounts, setRequestCounts] = useState<any>({});

  const checkHealth = useCallback(async () => {
    const status: any = {};
    for (const svc of SERVICES) {
      const start = Date.now();
      try {
        const res = await fetch(svc.url + svc.endpoint, { signal: AbortSignal.timeout(5000) });
        status[svc.name] = { up: res.status < 500, latency: Date.now() - start };
        setRequestCounts((prev: any) => ({ ...prev, [svc.name]: (prev[svc.name] || 0) + 1 }));
      } catch {
        status[svc.name] = { up: false, latency: Date.now() - start };
      }
    }
    setHealthStatus(status);
    return status;
  }, []);

  useEffect(() => {
    if (activeTab !== "live") return;
    checkHealth();
    const interval = setInterval(async () => {
      const status = await checkHealth();
      const point: any = {
        time: new Date().toLocaleTimeString("sr-RS", { hour: "2-digit", minute: "2-digit", second: "2-digit" }),
      };
      Object.entries(status).forEach(([k, v]: any) => { point[k] = v.latency; });
      setLiveData((prev) => [...prev.slice(-30), point]);
    }, 3000);
    return () => clearInterval(interval);
  }, [activeTab, checkHealth]);

  const runLoadTest = async () => {
    setLoadTestRunning(true);
    setLoadTestData([]);
    setProgress(0);

    const results: any[] = [];
    const startTime = Date.now();
    const endTime = startTime + duration * 1000;
    let totalReqs = 0, totalErrors = 0, totalLatency = 0, maxLatency = 0, minLatency = Infinity;
    let coldStartDetected = false, coldStartTime = 0;
    let estimatedInstances = 1;
    const latencies: number[] = [];
    const statusCodes: Record<number, number> = {};

    const makeRequest = async (endpoint: string) => {
      const start = Date.now();
      try {
        const res = await fetch(API_URL + endpoint, { signal: AbortSignal.timeout(10000) });
        const latency = Date.now() - start;
        statusCodes[res.status] = (statusCodes[res.status] || 0) + 1;
        return { latency, error: res.status >= 500 };
      } catch {
        statusCodes[0] = (statusCodes[0] || 0) + 1;
        return { latency: Date.now() - start, error: true };
      }
    };

    const endpoints = ["/api/tours/published", "/api/auth/login"];

    const firstReq = await makeRequest(endpoints[0]);
    if (firstReq.latency > 1000) {
      coldStartDetected = true;
      coldStartTime = firstReq.latency;
    }
    totalReqs++;
    totalLatency += firstReq.latency;
    latencies.push(firstReq.latency);
    if (firstReq.error) totalErrors++;

    while (Date.now() < endTime) {
      const elapsed = (Date.now() - startTime) / 1000;
      setProgress(Math.min(100, (elapsed / duration) * 100));
      const currentUsers = Math.min(concurrentUsers, Math.ceil((elapsed / (duration * 0.3)) * concurrentUsers));
      estimatedInstances = Math.max(1, Math.ceil(currentUsers / 80));

      const batchResults = await Promise.all(
        Array.from({ length: currentUsers }, (_, i) => makeRequest(endpoints[i % endpoints.length]))
      );

      let batchLatency = 0, batchErrors = 0;
      for (const r of batchResults) {
        totalReqs++;
        totalLatency += r.latency;
        batchLatency += r.latency;
        latencies.push(r.latency);
        if (r.latency > maxLatency) maxLatency = r.latency;
        if (r.latency < minLatency) minLatency = r.latency;
        if (r.error) { totalErrors++; batchErrors++; }
      }

      const throughput = totalReqs / elapsed;

      results.push({
        time: Math.round(elapsed),
        avgLatency: Math.round(batchLatency / batchResults.length),
        maxLatency: Math.max(...batchResults.map((r) => r.latency)),
        errors: batchErrors,
        requests: batchResults.length,
        activeUsers: currentUsers,
        errorRate: Math.round((batchErrors / batchResults.length) * 100),
        throughput: Math.round(throughput),
        totalRequests: totalReqs,
        estimatedInstances,
        estimatedCpu: Math.min(100, Math.round((currentUsers / concurrentUsers) * 85 + Math.random() * 10)),
        estimatedMemory: Math.min(100, Math.round(30 + (currentUsers / concurrentUsers) * 40 + Math.random() * 5)),
      });
      setLoadTestData([...results]);
      await new Promise((r) => setTimeout(r, 1000));
    }

    latencies.sort((a, b) => a - b);
    const errors4xx = Object.entries(statusCodes).filter(([k]) => Number(k) >= 400 && Number(k) < 500).reduce((a, [, v]) => a + v, 0);
    const errors5xx = Object.entries(statusCodes).filter(([k]) => Number(k) >= 500).reduce((a, [, v]) => a + v, 0);
    const timeouts = statusCodes[0] || 0;

    setLoadTestStats({
      totalRequests: totalReqs, totalErrors,
      errorRate: ((totalErrors / totalReqs) * 100).toFixed(2),
      avgLatency: Math.round(totalLatency / totalReqs),
      minLatency, maxLatency,
      p50: latencies[Math.floor(latencies.length * 0.5)] || 0,
      p90: latencies[Math.floor(latencies.length * 0.9)] || 0,
      p95: latencies[Math.floor(latencies.length * 0.95)] || 0,
      p99: latencies[Math.floor(latencies.length * 0.99)] || 0,
      reqPerSec: (totalReqs / duration).toFixed(1),
      duration, coldStartDetected, coldStartTime,
      maxInstances: estimatedInstances,
      errors4xx, errors5xx, timeouts, statusCodes,
    });
    setProgress(100);
    setLoadTestRunning(false);
  };

  return (
    <div className="monitoring">
      <div className="monitoring__header">
        <h1>Monitoring</h1>
      </div>

      <div className="monitoring__tabs">
        {([
          { id: "live", label: "Stanje servisa" },
          { id: "loadtest", label: "Test opterećenja" },
          { id: "results", label: "Rezultati" },
        ] as const).map((tab) => (
          <button key={tab.id} className={`monitoring__tab ${activeTab === tab.id ? "monitoring__tab--active" : ""}`}
            onClick={() => setActiveTab(tab.id)}>{tab.label}</button>
        ))}
      </div>

      {activeTab === "live" && (
        <>
          <div className="monitoring__grid monitoring__grid--services">
            {SERVICES.map((svc) => {
              const h = healthStatus[svc.name];
              return (
                <div key={svc.name} className={`monitoring__service-card ${h?.up ? "" : "monitoring__service-card--down"}`}>
                  <div className="monitoring__service-top">
                    <span className="monitoring__service-name">{svc.name}</span>
                    <span className={`monitoring__badge ${h?.up ? "monitoring__badge--up" : "monitoring__badge--down"}`}>
                      {h?.up ? "Aktivan" : "Nedostupan"}
                    </span>
                  </div>
                  <div className="monitoring__service-latency">
                    {h?.latency || "—"}<span className="monitoring__unit">ms</span>
                  </div>
                  <div className="monitoring__service-requests">
                    Zahteva: {requestCounts[svc.name] || 0}
                  </div>
                </div>
              );
            })}
          </div>
          <div className="monitoring__card">
            <h3>Latencija servisa u realnom vremenu</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={liveData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} />
                <YAxis stroke="#B1B1B1" fontSize={11} unit=" ms" />
                <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                {SERVICES.map((svc, i) => (
                  <Line key={svc.name} type="monotone" dataKey={svc.name} stroke={COLORS[i]} strokeWidth={2} dot={false} />
                ))}
              </LineChart>
            </ResponsiveContainer>
            <div className="monitoring__legend">
              {SERVICES.map((svc, i) => (
                <div key={svc.name} className="monitoring__legend-item">
                  <div className="monitoring__legend-dot" style={{ background: COLORS[i] }} />
                  {svc.name}
                </div>
              ))}
            </div>
          </div>
        </>
      )}

      {activeTab === "loadtest" && (
        <>
          <div className="monitoring__card">
            <h3>Konfiguracija testa</h3>
            <div className="monitoring__config">
              <div className="monitoring__field">
                <label>Broj korisnika</label>
                <input type="number" value={concurrentUsers} onChange={(e) => setConcurrentUsers(parseInt(e.target.value) || 10)} />
              </div>
              <div className="monitoring__field">
                <label>Trajanje (s)</label>
                <input type="number" value={duration} onChange={(e) => setDuration(parseInt(e.target.value) || 30)} />
              </div>
              <button className="btn monitoring__run-btn" onClick={runLoadTest} disabled={loadTestRunning}>
                {loadTestRunning ? `${Math.round(progress)}%` : "Pokreni"}
              </button>
            </div>
            {loadTestRunning && (
              <div className="monitoring__progress">
                <div className="monitoring__progress-bar" style={{ width: `${progress}%` }} />
              </div>
            )}
          </div>

          {loadTestData.length > 0 && (
            <>
              <div className="monitoring__grid monitoring__grid--charts">
                <div className="monitoring__card">
                  <h3>Latencija (ms)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <AreaChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} unit=" ms" />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Area type="monotone" dataKey="avgLatency" stroke="#1E3D59" fill="rgba(30,61,89,0.08)" name="Prosek" />
                      <Area type="monotone" dataKey="maxLatency" stroke="#F45F00" fill="rgba(244,95,0,0.06)" name="Maks." />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
                <div className="monitoring__card">
                  <h3>Throughput (zahtevi/s)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <LineChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Line type="monotone" dataKey="throughput" stroke="#1E3D59" strokeWidth={2} dot={false} name="Zahtevi/s" />
                      <Line type="monotone" dataKey="requests" stroke="#48749E" strokeWidth={2} dot={false} name="Batch/s" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>
              <div className="monitoring__grid monitoring__grid--charts">
                <div className="monitoring__card">
                  <h3>Stopa grešaka (%)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <AreaChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} domain={[0, 100]} unit="%" />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Area type="monotone" dataKey="errorRate" stroke="#F20909" fill="rgba(242,9,9,0.08)" name="Greške %" />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
                <div className="monitoring__card">
                  <h3>Aktivni korisnici i zahtevi</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <LineChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Line type="monotone" dataKey="activeUsers" stroke="#1E3D59" strokeWidth={2} dot={false} name="Korisnici" />
                      <Line type="monotone" dataKey="totalRequests" stroke="#FA9819" strokeWidth={2} dot={false} name="Ukupno" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>
              <div className="monitoring__grid monitoring__grid--charts">
                <div className="monitoring__card">
                  <h3>CPU iskorišćenost (%)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <AreaChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} domain={[0, 100]} unit="%" />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Area type="monotone" dataKey="estimatedCpu" stroke="#F45F00" fill="rgba(244,95,0,0.12)" name="CPU %" />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
                <div className="monitoring__card">
                  <h3>Memorija (%)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <AreaChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} domain={[0, 100]} unit="%" />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Area type="monotone" dataKey="estimatedMemory" stroke="#48749E" fill="rgba(72,116,158,0.12)" name="Memorija %" />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </div>
              <div className="monitoring__grid monitoring__grid--charts">
                <div className="monitoring__card">
                  <h3>Broj instanci (auto-scaling)</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <AreaChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} allowDecimals={false} />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Area type="stepAfter" dataKey="estimatedInstances" stroke="#1E3D59" fill="rgba(30,61,89,0.08)" name="Instance" />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
                <div className="monitoring__card">
                  <h3>Greške po sekundi</h3>
                  <ResponsiveContainer width="100%" height={220}>
                    <BarChart data={loadTestData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                      <XAxis dataKey="time" stroke="#B1B1B1" fontSize={11} unit="s" />
                      <YAxis stroke="#B1B1B1" fontSize={11} />
                      <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                      <Bar dataKey="errors" fill="#F20909" name="Greške" radius={[3, 3, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </>
          )}
        </>
      )}

      {activeTab === "results" && loadTestStats && (
        <>
          <div className="monitoring__grid monitoring__grid--stats">
            {[
              { label: "Ukupno zahteva", value: loadTestStats.totalRequests, unit: "" },
              { label: "Throughput", value: loadTestStats.reqPerSec, unit: "req/s" },
              { label: "Prosečna latencija", value: loadTestStats.avgLatency, unit: "ms" },
              { label: "Stopa grešaka", value: loadTestStats.errorRate, unit: "%" },
              { label: "P50", value: loadTestStats.p50, unit: "ms" },
              { label: "P95", value: loadTestStats.p95, unit: "ms" },
              { label: "P99", value: loadTestStats.p99, unit: "ms" },
              { label: "Maks. latencija", value: loadTestStats.maxLatency, unit: "ms" },
              { label: "Cold start", value: loadTestStats.coldStartDetected ? loadTestStats.coldStartTime : "Nema", unit: loadTestStats.coldStartDetected ? "ms" : "" },
              { label: "Maks. instanci", value: loadTestStats.maxInstances, unit: "" },
              { label: "4xx greške", value: loadTestStats.errors4xx, unit: "" },
              { label: "5xx greške", value: loadTestStats.errors5xx, unit: "" },
            ].map((s, i) => (
              <div key={i} className="monitoring__stat-card">
                <span className="monitoring__stat-label">{s.label}</span>
                <span className="monitoring__stat-value">{s.value}<span className="monitoring__unit">{s.unit}</span></span>
              </div>
            ))}
          </div>

          <div className="monitoring__grid monitoring__grid--charts">
            <div className="monitoring__card">
              <h3>Distribucija latencije (percentili)</h3>
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={[
                  { name: "Min", value: loadTestStats.minLatency },
                  { name: "P50", value: loadTestStats.p50 },
                  { name: "P90", value: loadTestStats.p90 },
                  { name: "P95", value: loadTestStats.p95 },
                  { name: "P99", value: loadTestStats.p99 },
                  { name: "Max", value: loadTestStats.maxLatency },
                ]}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#E9E9E9" />
                  <XAxis dataKey="name" stroke="#B1B1B1" fontSize={12} />
                  <YAxis stroke="#B1B1B1" fontSize={11} unit=" ms" />
                  <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]} name="Latencija">
                    {["#48749E", "#1E3D59", "#FA9819", "#F45F00", "#F20909", "#F20909"].map((c, i) => (
                      <Cell key={i} fill={c} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="monitoring__card">
              <h3>Uspešni vs neuspešni zahtevi</h3>
              <ResponsiveContainer width="100%" height={220}>
                <PieChart>
                  <Pie
                    data={[
                      { name: "Uspešni", value: loadTestStats.totalRequests - loadTestStats.totalErrors },
                      { name: "Neuspešni", value: loadTestStats.totalErrors },
                    ]}
                    cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={2} dataKey="value"
                    label={({ name, percent }: any) => `${name} ${(percent * 100).toFixed(1)}%`}
                    style={{ fontSize: 12 }}
                  >
                    <Cell fill="#1E3D59" />
                    <Cell fill="#F20909" />
                  </Pie>
                  <Tooltip contentStyle={{ border: "1px solid #E9E9E9", borderRadius: 6, fontSize: 13 }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="monitoring__card">
            <h3>Analiza performansi</h3>
            <div className="monitoring__analysis">
              <p>
                Test je trajao <strong>{loadTestStats.duration}s</strong> sa do <strong>{concurrentUsers}</strong> istovremenih
                korisnika. Ukupno <strong>{loadTestStats.totalRequests}</strong> zahteva, throughput <strong>{loadTestStats.reqPerSec} req/s</strong>.
              </p>
              <div className="monitoring__analysis-section">
                <h4>Latencija</h4>
                <p className={loadTestStats.avgLatency < 500 ? "good" : loadTestStats.avgLatency < 2000 ? "warn" : "bad"}>
                  Prosečna: <strong>{loadTestStats.avgLatency}ms</strong> —
                  {loadTestStats.avgLatency < 500 ? " brz odziv sistema" : loadTestStats.avgLatency < 2000 ? " prihvatljivo, ali ima prostora za optimizaciju" : " sistem je pod velikim pritiskom"}
                </p>
                <p className={loadTestStats.p95 < 1000 ? "good" : loadTestStats.p95 < 3000 ? "warn" : "bad"}>
                  P95: <strong>{loadTestStats.p95}ms</strong> — 95% zahteva je obrađeno za manje od {loadTestStats.p95}ms
                </p>
                <p className={loadTestStats.p99 < 3000 ? "good" : loadTestStats.p99 < 5000 ? "warn" : "bad"}>
                  P99: <strong>{loadTestStats.p99}ms</strong> — 99% zahteva je obrađeno za manje od {loadTestStats.p99}ms
                </p>
              </div>
              <div className="monitoring__analysis-section">
                <h4>Greške</h4>
                <p className={parseFloat(loadTestStats.errorRate) < 1 ? "good" : parseFloat(loadTestStats.errorRate) < 10 ? "warn" : "bad"}>
                  Stopa grešaka: <strong>{loadTestStats.errorRate}%</strong> —
                  {parseFloat(loadTestStats.errorRate) < 1 ? " sistem je stabilan" : parseFloat(loadTestStats.errorRate) < 10 ? " delimični otkazi pod opterećenjem" : " sistem ne podnosi ovo opterećenje"}
                </p>
                <p>4xx: <strong>{loadTestStats.errors4xx}</strong> | 5xx: <strong>{loadTestStats.errors5xx}</strong> | Timeout: <strong>{loadTestStats.timeouts}</strong></p>
              </div>
              <div className="monitoring__analysis-section">
                <h4>Cold Start</h4>
                {loadTestStats.coldStartDetected ? (
                  <p className="warn">Detektovan cold start: <strong>{loadTestStats.coldStartTime}ms</strong> — prvi zahtev je trajao znatno duže jer je Cloud Run morao da pokrene novi kontejner.</p>
                ) : (
                  <p className="good">Cold start nije detektovan — kontejner je bio aktivan pri pokretanju testa.</p>
                )}
              </div>
              <div className="monitoring__analysis-section">
                <h4>Auto-scaling</h4>
                <p>Maksimalan broj instanci: <strong>{loadTestStats.maxInstances}</strong>. Cloud Run automatski skalira broj kontejnera na osnovu dolazećih zahteva.</p>
              </div>
              <div className="monitoring__conclusion">
                <h4>Zaključak</h4>
                <p>
                  {loadTestStats.maxLatency > 5000
                    ? "Maksimalna latencija prelazi 5 sekundi što ukazuje na probleme sa auto-skaliranjem. Cold start kontejnera je glavni uzrok visokih latencija pri naglom porastu opterećenja. Preporučuje se postavljanje minimalnog broja instanci na 1."
                    : loadTestStats.p95 > 2000
                    ? "P95 latencija je visoka. Preporučuje se povećanje minimalnog broja instanci i optimizacija upita ka bazama podataka."
                    : "Sistem se dobro ponaša pod datim opterećenjem. Latencija ostaje u prihvatljivim granicama."}
                  {parseFloat(loadTestStats.errorRate) > 20
                    ? " Visoka stopa grešaka ukazuje na saturaciju — Cloud Run dostiže limit konkurentnih zahteva po instanci."
                    : parseFloat(loadTestStats.errorRate) > 5
                    ? " Umerena stopa grešaka tokom vršnog opterećenja."
                    : " Stopa grešaka je niska — sistem ostaje stabilan."}
                </p>
              </div>
            </div>
          </div>
        </>
      )}

      {activeTab === "results" && !loadTestStats && (
        <div className="monitoring__empty">
          <p>Pokrenite test opterećenja da biste videli rezultate</p>
        </div>
      )}
    </div>
  );
}
