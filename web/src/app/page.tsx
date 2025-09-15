import Link from 'next/link'

export default function LandingPage() {
  return (
    <main>
      <section className="container-prose pt-28 pb-16">
        <div className="grid md:grid-cols-2 gap-10 items-center">
          <div>
            <h1 className="text-4xl md:text-6xl font-semibold leading-tight">
              Dining, reimagined.
            </h1>
            <p className="mt-5 text-lg text-white/70 max-w-xl">
              Order from your table with a clean, minimal interface. Secure phone + OTP auth. Real-time status updates from the kitchen.
            </p>
            <div className="mt-8 flex gap-4">
              <Link href="/app" className="btn-primary">Start from your table</Link>
              <a href="#features" className="inline-flex items-center text-white/80 hover:text-white">Learn more →</a>
            </div>
          </div>
          <div className="card">
            <div className="aspect-[4/3] w-full rounded-lg bg-gradient-to-br from-white/10 to-white/5" />
          </div>
        </div>
      </section>

      <section id="features" className="container-prose py-16">
        <h2 className="text-2xl md:text-3xl font-semibold">Features</h2>
        <div className="mt-8 grid md:grid-cols-3 gap-6">
          {[
            { title: 'Fast ordering', desc: 'No apps to install. Scan, order, relax.' },
            { title: 'Secure login', desc: 'Phone number + OTP with authorized devices.' },
            { title: 'Live updates', desc: 'See when your order is being prepared and ready.' },
          ].map((f) => (
            <div key={f.title} className="card">
              <h3 className="text-lg font-semibold">{f.title}</h3>
              <p className="mt-2 text-white/70">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>
    </main>
  )
}



