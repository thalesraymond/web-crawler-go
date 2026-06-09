package indexer

import "testing"

func BenchmarkExtractPageTokens(b *testing.B) {
	htmlBody := `<html>
	<head>
		<title>Hello World</title>
		<script>var x = 1; console.log(x);</script>
		<style>body { color: red; } .test { font-size: 12px; }</style>
	</head>
	<body>
		<header><nav>Navigation</nav>Header text</header>
		<main>
			<h1>Main title</h1>
			<p>Here is some paragraph text with <b>bold</b> and <i>italic</i> words. <a href="https://example.com" class="link" id="link-1">Link text</a></p>
			<div>
				More text in a div.
			</div>
		</main>
		<aside>Aside text</aside>
		<footer>Footer text</footer>
	</body>
	</html>`

	for i := 0; i < b.N; i++ {
		ExtractPageTokens(htmlBody)
	}
}
