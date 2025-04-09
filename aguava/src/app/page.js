import Link from "next/link";
import Image from "next/image"

export default function Home() {
  return (
    <div>
      <h1>Aguava</h1>
      <div>
        <Image
           src="/images/cuandosueno.jpg"
           alt="Cuando Sueno"
           width={500}
           height={500}
           layout="intrinsic"
        />
      </div>
      <nav>
        <ul>
          <li><Link href="/releases">Go to Releases</Link></li>
          <li><Link href="/press">Go to Press</Link></li>
          <li><Link href="/contact">Go to Contact</Link></li>
        </ul>
      </nav>
    </div>
  );
}