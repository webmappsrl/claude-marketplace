#!/usr/bin/env perl
# Dal testo di un file agli identificatori che contengono una parola italiana.
#
# Si invoca con il path (serve solo per l'estensione) e il testo su stdin:
#
#   perl identifier-scan.pl path/del/file.ts < testo
#
# Scrive una riga `parola<TAB>identificatore` per ogni rilievo. Lo chiama
# identifier-check.sh, che ci mette intorno la lettura dell'evento e il messaggio.
#
# I due file di dati arrivano dalle variabili d'ambiente WORDS e ALLOWLIST.

use strict;
use warnings;

my $path = shift @ARGV // '';
my $src = do { local $/; <STDIN> };
$src = '' unless defined $src;

# ── il dizionario e le eccezioni ────────────────────────────────────────────
my (%bad, %allowed);
for my $spec ([$ENV{WORDS} // '', \%bad], [$ENV{ALLOWLIST} // '', \%allowed]) {
  my ($file, $into) = @$spec;
  next unless $file ne '' && -f $file;
  open my $fh, '<', $file or next;
  while (my $line = <$fh>) {
    $line =~ s/#.*//;            # il commento del motivo, sulla stessa riga
    $line =~ s/^\s+|\s+$//g;
    $into->{ lc $line } = 1 if $line ne '';
  }
  close $fh;
}

# ── di un markdown si guardano solo i blocchi di codice ─────────────────────
# È dove il piano scrive gli identificatori, prima ancora che il codice esista: se il
# nome sbagliato nasce lì, l'implementazione lo copia.
#
# Contano solo i blocchi di un linguaggio di programmazione: un blocco ```markdown o
# senza linguaggio, in un SKILL.md o in un overview, è un esempio di documento scritto
# in italiano, e segnalarlo riempirebbe di rumore ogni modifica alla documentazione.
my %code_fence = map { $_ => 1 } qw(
  php js jsx ts tsx javascript typescript go golang python py ruby rb java kotlin swift
  c cpp csharp cs rust sql graphql scss css
);
if ($path =~ /\.(md|markdown)$/i) {
  my @blocks;
  while ($src =~ /^[ \t]*```([A-Za-z+#]*)[^\n]*\n(.*?)^[ \t]*```/gms) {
    push @blocks, $2 if $code_fence{ lc $1 };
  }
  $src = join "\n", @blocks;
}

# ── commenti e stringhe fuori ───────────────────────────────────────────────
# L'italiano lì è prescritto, non è un difetto: un commento, un docblock, un messaggio
# per il dev e una stringa traducibile restano in italiano.
$src =~ s{/\*.*?\*/}{ }gs;                   # /* ... */
$src =~ s{(^|[^:])//[^\n]*}{$1 }g;           # // ...  (il [^:] salva http://)
$src =~ s{^\s*\#[^\n]*}{ }gm;                # # ...   (php, python, shell, yaml)
$src =~ s{--[^\n]*}{ }g if $path =~ /\.sql$/i;

# Il testo dentro un elemento JSX è una frase per chi legge lo schermo, non un nome:
# `<Text>Nessun overview sotto docs/features/</Text>`. Si toglie solo quando fra i due
# segni non c'è niente che somigli a codice, così `a > b && c < d` resta intero.
$src =~ s{>([^<>{}()=;&|]*)<}{> <}g if $path =~ /\.(tsx|jsx)$/i;

# I literal regex: `/^>\s*Stato:\s*(.+?)$/m` non dichiara niente che si chiami `Stato`.
# Vanno tolti prima delle stringhe, perché dentro una regex ci stanno apici e backtick.
$src =~ s{(^|[=(,:\[!&|?{};]|\breturn|\bcase)(\s*)/(?:\\.|\[(?:\\.|[^\]\\])*\]|[^/\\\n])+/[gimsuydv]*}{$1$2 }g;

# Due stringhe che però nominano un campo, e la regola le include esplicitamente:
# la chiave di un array — `'prezzo_totale' => 0` — e la colonna di una migration —
# `$table->string('prezzo')`. Si marcano prima di buttare via le stringhe, e restano
# riconoscibili dagli apici nel rilievo.
my @keys;
while ($src =~ /(['"])([A-Za-z_][A-Za-z0-9_]{3,})\1\s*=>/g)                       { push @keys, $2 }
while ($src =~ /->\s*[A-Za-z]+\s*\(\s*(['"])([A-Za-z_][A-Za-z0-9_]{3,})\1\s*\)/g) { push @keys, $2 }

$src =~ s{"(?:\\.|[^"\\])*"}{ }gs;           # "..."
$src =~ s{'(?:\\.|[^'\\])*'}{ }gs;           # '...'
$src =~ s{`(?:\\.|[^`\\])*`}{ }gs;           # `...`

# ── gli identificatori, spezzati in token ───────────────────────────────────
# Le chiavi si riportano fra apici, come sono scritte: così si distingue a colpo un
# `'prezzo'` che è una colonna da un `$prezzo` che è una variabile.
my @idents = ($src =~ /([A-Za-z_\$][A-Za-z0-9_]{2,})/g);
push @idents, map { "'$_'" } @keys;

my %seen;
for my $ident (@idents) {
  next if $seen{$ident}++;

  my $bare = $ident;
  $bare =~ s/^\$//;
  $bare =~ s/^'|'$//g;

  # camelCase, PascalCase, snake_case, SCREAMING_CASE: tutto a token minuscoli.
  my $spaced = $bare;
  $spaced =~ s/([a-z0-9])([A-Z])/$1 $2/g;
  $spaced =~ s/([A-Z]+)([A-Z][a-z])/$1 $2/g;
  $spaced =~ s/[_\-]+/ /g;

  for my $token (map { lc } split /\s+/, $spaced) {
    next if length($token) < 4;          # sotto le quattro lettere fa solo rumore
    next if $allowed{$token};
    print "$token\t$ident\n" if $bad{$token};
  }
}
