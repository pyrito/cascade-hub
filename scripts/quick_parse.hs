import System.Environment
import System.IO
import Text.Read

fix :: [Maybe Int] -> [Rational]
fix [] = []
fix (Nothing:xs) = fix xs
fix (Just a:xs) = toRational a/1000000000:fix xs

mean :: (Real a, Fractional b) => [a] -> b
mean xs = realToFrac (sum xs) / fromIntegral (length xs)

runParse :: [String] -> [Rational]
runParse x = fix (fmap (readMaybe :: String -> Maybe Int) x)

variance :: (Real a, Fractional b) => [a] -> b
variance x = let y = fmap realToFrac x in
             let n = realToFrac . length $ x in
             let s = fmap (mean x -) y in
             let e = fmap (^^2) s in
             sum e /(n-1)

sdev :: (Real a, Floating b) => [a] -> b
sdev = sqrt . variance

sder :: (Real a, Floating b) => [a] -> b
sder x = sqrt (variance x / (realToFrac . length $ x))

main = do
          args <- getArgs
          handle <- openFile (head args) ReadMode
          content <- hGetContents handle
          print . mean . runParse . lines $ content
          print . sder . runParse . lines $ content
          print . variance . runParse . lines $ content
          print . sdev . runParse . lines $ content
